// Package gateway serves registered media through a loopback-only endpoint.
package gateway

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

type resource struct {
	primary, backup string
	playlist        bool
}
type proxySession struct {
	resources map[resource]string
}
type Gateway struct {
	listener            net.Listener
	server              *http.Server
	client              *http.Client
	playlistReadTimeout time.Duration
	mu                  sync.RWMutex
	urls                map[string]resource
	owners              map[string]string
	proxies             map[string]*proxySession
	muxes               map[string]*muxSession
	closed              bool
}

func New() (*Gateway, error) {
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = 15 * time.Second
	g := &Gateway{listener: l, client: &http.Client{Transport: transport}, playlistReadTimeout: 15 * time.Second, urls: map[string]resource{}, owners: map[string]string{}, proxies: map[string]*proxySession{}, muxes: map[string]*muxSession{}}
	g.server = &http.Server{Handler: http.HandlerFunc(g.serve), ReadHeaderTimeout: 5 * time.Second}
	go g.server.Serve(l)
	return g, nil
}
func (g *Gateway) Register(raw string) (string, error) { return g.RegisterStream(raw, "", false) }
func (g *Gateway) RegisterStream(raw, backup string, playlist bool) (string, error) {
	return g.registerStream("", raw, backup, playlist)
}

// registerStream starts a proxy session when parent is empty. Child resources
// belong to the same session as their parent playlist.
func (g *Gateway) registerStream(parent, raw, backup string, playlist bool) (string, error) {
	if raw == "" {
		return "", fmt.Errorf("无效的媒体地址")
	}
	for _, value := range []string{raw, backup} {
		if value == "" && value == backup {
			continue
		}
		u, err := url.Parse(value)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.User != nil {
			return "", fmt.Errorf("无效的媒体地址")
		}
	}
	item := resource{raw, backup, playlist}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return "", fmt.Errorf("媒体网关已关闭")
	}
	var session *proxySession
	owner := ""
	if parent != "" {
		owner = g.owners[parent]
		session = g.proxies[owner]
		if session == nil {
			return "", fmt.Errorf("播放资源已释放")
		}
	}
	token := ""
	if session != nil {
		token = session.resources[item]
	}
	if token == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		token = hex.EncodeToString(b)
		if session == nil {
			owner = token
			session = &proxySession{resources: make(map[resource]string)}
			g.proxies[owner] = session
		}
		session.resources[item] = token
		g.urls[token] = item
		g.owners[token] = owner
	}
	suffix := ""
	if playlist {
		suffix = "/index.m3u8"
	}
	return "http://" + g.listener.Addr().String() + "/" + token + suffix, nil
}
func (g *Gateway) Close() error {
	g.mu.Lock()
	g.closed = true
	sessions := g.muxes
	g.muxes = make(map[string]*muxSession)
	g.urls = make(map[string]resource)
	g.owners = make(map[string]string)
	g.proxies = make(map[string]*proxySession)
	g.mu.Unlock()
	for _, session := range sessions {
		session.stop()
	}
	g.client.CloseIdleConnections()
	return g.server.Close()
}

var uriAttribute = regexp.MustCompile(`URI="([^"]+)"`)

func (g *Gateway) rewrite(body []byte, base, backupBase *url.URL, parent string) ([]byte, error) {
	var out bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	rewriteURL := func(value string, forcePlaylist bool) (string, error) {
		u, err := url.Parse(value)
		if err != nil {
			return "", err
		}
		resolved := base.ResolveReference(u)
		playlist := forcePlaylist || strings.HasSuffix(strings.ToLower(resolved.Path), ".m3u8")
		backup := ""
		// Absolute and network-path references may point at an unrelated service.
		// Only map relative references to the other CDN's playlist base.
		if backupBase != nil && !u.IsAbs() && u.Host == "" {
			backup = backupBase.ResolveReference(u).String()
		}
		return g.registerStream(parent, resolved.String(), backup, playlist)
	}
	nextPlaylist := false
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			v, err := rewriteURL(trimmed, nextPlaylist)
			if err != nil {
				return nil, err
			}
			line = v
			nextPlaylist = false
		} else if strings.HasPrefix(trimmed, "#") {
			if strings.HasPrefix(trimmed, "#EXT-X-STREAM-INF:") {
				nextPlaylist = true
			}
			var rewriteErr error
			line = uriAttribute.ReplaceAllStringFunc(line, func(s string) string {
				if rewriteErr != nil {
					return s
				}
				parts := uriAttribute.FindStringSubmatch(s)
				isPlaylist := strings.HasPrefix(trimmed, "#EXT-X-MEDIA:") || strings.HasPrefix(trimmed, "#EXT-X-I-FRAME-STREAM-INF:") || strings.HasPrefix(trimmed, "#EXT-X-RENDITION-REPORT:")
				v, err := rewriteURL(parts[1], isPlaylist)
				if err != nil {
					rewriteErr = err
					return s
				}
				return `URI="` + v + `"`
			})
			if rewriteErr != nil {
				return nil, rewriteErr
			}
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}
	return out.Bytes(), scanner.Err()
}

func (g *Gateway) preflight(w http.ResponseWriter, r *http.Request, origin string) {
	method := r.Header.Get("Access-Control-Request-Method")
	if origin == "" || (method != "GET" && method != "HEAD") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	for _, header := range strings.Split(r.Header.Get("Access-Control-Request-Headers"), ",") {
		if name := strings.TrimSpace(header); name != "" && !strings.EqualFold(name, "range") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD")
	w.Header().Set("Access-Control-Allow-Headers", "Range")
	if strings.EqualFold(r.Header.Get("Access-Control-Request-Private-Network"), "true") {
		w.Header().Set("Access-Control-Allow-Private-Network", "true")
	}
	w.Header().Set("Vary", "Origin, Access-Control-Request-Method, Access-Control-Request-Headers, Access-Control-Request-Private-Network")
	w.WriteHeader(http.StatusNoContent)
}

func validPlaylist(data []byte) ([]byte, bool) {
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	first, _, _ := bytes.Cut(data, []byte("\n"))
	return data, bytes.Equal(bytes.TrimSuffix(first, []byte("\r")), []byte("#EXTM3U"))
}

func (g *Gateway) serve(w http.ResponseWriter, r *http.Request) {
	if r.Host != g.listener.Addr().String() || (r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS") {
		http.Error(w, "forbidden", 403)
		return
	}
	origin := r.Header.Get("Origin")
	if origin != "" && origin != "http://wails.localhost" && origin != "https://wails.localhost" && origin != "http://localhost:34115" {
		http.Error(w, "forbidden", 403)
		return
	}
	token := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")[0]
	g.mu.RLock()
	item, ok := g.urls[token]
	mux := g.muxes[token]
	g.mu.RUnlock()
	if !ok && mux == nil {
		http.NotFound(w, r)
		return
	}
	if r.Method == "OPTIONS" {
		g.preflight(w, r, origin)
		return
	}
	if mux != nil {
		g.serveMux(w, r, token, mux, origin)
		return
	}
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range, Accept-Ranges")
	}
	var resp *http.Response
	var err error
	var rewritten []byte
	for i, raw := range []string{item.primary, item.backup} {
		if raw == "" {
			continue
		}
		ctx, cancel := context.WithCancel(r.Context())
		req, e := http.NewRequestWithContext(ctx, r.Method, raw, nil)
		if e != nil {
			cancel()
			err = e
			continue
		}
		if v := r.Header.Get("Range"); v != "" {
			req.Header.Set("Range", v)
		}
		resp, err = g.client.Do(req)
		if err != nil || resp.StatusCode >= 500 || (resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != http.StatusRequestedRangeNotSatisfiable) {
			if resp != nil {
				resp.Body.Close()
				resp = nil
			}
			cancel()
			continue
		}
		contentType := strings.ToLower(resp.Header.Get("Content-Type"))
		isPlaylist := item.playlist || strings.Contains(contentType, "mpegurl")
		if isPlaylist && r.Method == http.MethodGet && resp.StatusCode == http.StatusOK {
			timer := time.AfterFunc(g.playlistReadTimeout, cancel)
			body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024+1))
			timer.Stop()
			tooLarge := len(body) > 4*1024*1024
			body, valid := validPlaylist(body)
			if readErr == nil && !tooLarge && valid {
				alternate := item.backup
				if i == 1 {
					alternate = item.primary
				}
				var alternateBase *url.URL
				if alternate != "" {
					alternateBase, _ = url.Parse(alternate)
				}
				rewritten, readErr = g.rewrite(body, resp.Request.URL, alternateBase, token)
			}
			if readErr != nil || tooLarge || !valid {
				resp.Body.Close()
				resp = nil
				cancel()
				continue
			}
		}
		defer cancel()
		break
	}
	if err != nil || resp == nil {
		http.Error(w, "upstream unavailable", 502)
		return
	}
	defer resp.Body.Close()
	if rewritten != nil {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		_, _ = w.Write(rewritten)
		return
	}
	for _, h := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges"} {
		if v := resp.Header.Get(h); v != "" {
			w.Header().Set(h, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
