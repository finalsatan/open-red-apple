package gateway

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// muxSession owns one ffmpeg process and its temporary HLS output.
type muxSession struct {
	dir      string
	cancel   context.CancelFunc
	done     chan struct{}
	err      error // read only after done is closed
	stopOnce sync.Once
	cleanup  func()
	mu       sync.RWMutex
	stopped  bool
}

func (s *muxSession) stop() {
	s.stopOnce.Do(func() {
		s.cancel()
		<-s.done
		s.mu.Lock()
		s.stopped = true
		_ = os.RemoveAll(s.dir)
		s.mu.Unlock()
		s.cleanup()
	})
}

func findFFmpeg() (string, error) {
	name := "ffmpeg"
	if filepath.Separator == '\\' {
		name += ".exe"
	}
	if exe, err := os.Executable(); err == nil {
		bundled := filepath.Join(filepath.Dir(exe), name)
		if info, err := os.Stat(bundled); err == nil && info.Mode().IsRegular() {
			return bundled, nil
		}
	}
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("播放独立音轨需要 FFmpeg，请将 %s 放在程序目录或加入 PATH", name)
}

// RegisterMux starts a clear video+audio remux and waits for its first segment.
// The process reads through the gateway so video requests also get CDN fallback.
func (g *Gateway) RegisterMux(ctx context.Context, video, backup, audio string, videoIsHLS bool, durationSeconds int) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	binary, err := findFFmpeg()
	if err != nil {
		return "", err
	}
	if audio == "" {
		return "", fmt.Errorf("此分集缺少独立音轨")
	}
	videoURL, err := g.RegisterStream(video, backup, videoIsHLS)
	if err != nil {
		return "", err
	}
	removeResource := g.Release
	audioURL, err := g.RegisterStream(audio, "", false)
	if err != nil {
		removeResource(videoURL)
		return "", err
	}
	dir, err := os.MkdirTemp("", "win-red-apple-mux-*")
	if err != nil {
		removeResource(videoURL)
		removeResource(audioURL)
		return "", fmt.Errorf("创建播放缓存失败: %w", err)
	}
	var random [32]byte
	if _, err = rand.Read(random[:]); err != nil {
		_ = os.RemoveAll(dir)
		removeResource(videoURL)
		removeResource(audioURL)
		return "", err
	}
	token := hex.EncodeToString(random[:])
	runCtx, cancel := context.WithCancel(context.Background())
	session := &muxSession{dir: dir, cancel: cancel, done: make(chan struct{}), cleanup: func() { removeResource(videoURL); removeResource(audioURL) }}
	args := []string{"-nostdin", "-hide_banner", "-loglevel", "error", "-i", videoURL, "-i", audioURL, "-map", "0:v:0", "-map", "1:a:0", "-c", "copy", "-f", "hls", "-hls_time", "4", "-hls_playlist_type", "event", "-hls_segment_type", "mpegts", "-hls_flags", "independent_segments+temp_file", "-hls_segment_filename", filepath.Join(dir, "segment%06d.ts"), filepath.Join(dir, "index.m3u8")}
	cmd := exec.CommandContext(runCtx, binary, args...)
	hideProcessWindow(cmd)
	if err = ctx.Err(); err != nil {
		cancel()
		_ = os.RemoveAll(dir)
		session.cleanup()
		return "", err
	}
	if err = cmd.Start(); err != nil {
		cancel()
		_ = os.RemoveAll(dir)
		session.cleanup()
		return "", fmt.Errorf("启动 FFmpeg 失败: %w", err)
	}
	go func() { session.err = cmd.Wait(); close(session.done) }()
	g.mu.Lock()
	if g.closed {
		g.mu.Unlock()
		session.stop()
		return "", fmt.Errorf("媒体网关已关闭")
	}
	g.muxes[token] = session
	g.mu.Unlock()
	cleanup := func() { g.releaseMux(token, session) }
	timeout := time.NewTimer(60 * time.Second)
	defer timeout.Stop()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-session.done:
			if session.err != nil {
				cleanup()
				return "", fmt.Errorf("FFmpeg 无法合并视频和音轨，请检查资源或 FFmpeg 的格式支持")
			}
		default:
		}
		// A playlist may exist before its first segment is ready.
		data, _ := os.ReadFile(filepath.Join(dir, "index.m3u8"))
		if strings.Contains(string(data), "#EXTINF:") {
			source := "http://" + g.listener.Addr().String() + "/" + token + "/index.m3u8?mux=1"
			if durationSeconds > 0 {
				source += "&duration=" + strconv.Itoa(durationSeconds)
			}
			return source, nil
		}
		select {
		case <-ctx.Done():
			cleanup()
			return "", ctx.Err()
		case <-session.done:
			data, _ := os.ReadFile(filepath.Join(dir, "index.m3u8"))
			if session.err == nil && strings.Contains(string(data), "#EXTINF:") {
				continue
			}
			cleanup()
			return "", fmt.Errorf("FFmpeg 无法合并视频和音轨，请检查资源或 FFmpeg 的格式支持")
		case <-timeout.C:
			cleanup()
			return "", fmt.Errorf("准备视频超时，请检查网络后重试")
		case <-ticker.C:
		}
	}
}

func (g *Gateway) releaseMux(token string, expected *muxSession) {
	g.mu.Lock()
	if g.muxes[token] == expected {
		delete(g.muxes, token)
	}
	g.mu.Unlock()
	expected.stop()
}

// Release stops an owned remux or proxy and releases its child resources.
func (g *Gateway) Release(raw string) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "http" || u.Host != g.listener.Addr().String() {
		return
	}
	token := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")[0]
	g.mu.Lock()
	session := g.muxes[token]
	if g.owners[token] == token {
		if proxy := g.proxies[token]; proxy != nil {
			for _, child := range proxy.resources {
				delete(g.urls, child)
				delete(g.owners, child)
			}
			delete(g.proxies, token)
		}
	}
	g.mu.Unlock()
	if session != nil {
		g.releaseMux(token, session)
	}
}

var segmentName = regexp.MustCompile(`^segment[0-9]{6}\.ts$`)

func (g *Gateway) serveMux(w http.ResponseWriter, r *http.Request, token string, session *muxSession, origin string) {
	session.mu.RLock()
	defer session.mu.RUnlock()
	if session.stopped {
		http.NotFound(w, r)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/"+token+"/")
	if name != "index.m3u8" && !segmentName.MatchString(name) {
		http.NotFound(w, r)
		return
	}
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range, Accept-Ranges")
	}
	if name == "index.m3u8" {
		select {
		case <-session.done:
			if session.err != nil && !errors.Is(session.err, context.Canceled) {
				http.Error(w, "media remux failed", http.StatusBadGateway)
				return
			}
		default:
		}
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Header().Set("Cache-Control", "no-store")
	} else {
		w.Header().Set("Content-Type", "video/mp2t")
	}
	http.ServeFile(w, r, filepath.Join(session.dir, name))
}
