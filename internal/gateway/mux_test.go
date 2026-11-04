package gateway

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMuxClearSeparateTracks(t *testing.T) {
	if testing.Short() {
		t.Skip("ffmpeg integration test")
	}
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg unavailable")
	}
	dir := t.TempDir()
	makeTrack := func(args ...string) {
		t.Helper()
		cmd := exec.Command(ffmpeg, append([]string{"-hide_banner", "-loglevel", "error", "-y"}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("create test media: %v: %s", err, out)
		}
	}
	makeTrack("-f", "lavfi", "-i", "testsrc2=size=160x90:rate=10", "-t", "3", "-c:v", "libx264", "-threads", "1", "-g", "10", filepath.Join(dir, "video.mp4"))
	makeTrack("-f", "lavfi", "-i", "sine=frequency=440", "-t", "3", "-c:a", "aac", filepath.Join(dir, "audio.m4a"))
	upstream := httptest.NewServer(http.FileServer(http.Dir(dir)))
	defer upstream.Close()
	g, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	source, err := g.RegisterMux(ctx, upstream.URL+"/video.mp4", "", upstream.URL+"/audio.m4a", false, 3)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(source)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Query().Get("duration") != "3" {
		t.Fatalf("duration metadata: %s", parsed.RawQuery)
	}
	preflight, _ := http.NewRequest(http.MethodOptions, source, nil)
	preflight.Header.Set("Origin", "http://wails.localhost")
	preflight.Header.Set("Access-Control-Request-Method", "GET")
	preflightResponse, err := http.DefaultClient.Do(preflight)
	if err != nil {
		t.Fatal(err)
	}
	preflightResponse.Body.Close()
	if preflightResponse.StatusCode != http.StatusNoContent {
		t.Fatalf("mux preflight: %s", preflightResponse.Status)
	}
	token := strings.Split(strings.TrimPrefix(parsed.Path, "/"), "/")[0]
	g.mu.RLock()
	session := g.muxes[token]
	g.mu.RUnlock()
	if session == nil {
		t.Fatal("session not registered")
	}
	select {
	case <-session.done:
		if session.err != nil {
			t.Fatalf("ffmpeg: %v", session.err)
		}
	case <-ctx.Done():
		t.Fatal("mux did not finish")
	}
	resp, err := http.Get(source)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 || !strings.Contains(string(body), "#EXT-X-ENDLIST") {
		t.Fatalf("playlist %s: %s", resp.Status, body)
	}
	if ffprobe, err := exec.LookPath("ffprobe"); err == nil {
		cmd := exec.Command(ffprobe, "-v", "error", "-show_entries", "stream=codec_type", "-of", "csv=p=0", source)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("ffprobe: %v: %s", err, out)
		}
		if !strings.Contains(string(out), "video") || !strings.Contains(string(out), "audio") {
			t.Fatalf("missing track: %s", out)
		}
	}
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/%s/../../video.mp4", parsed.Host, token), nil)
	req.URL.Path = "/" + token + "/../../video.mp4"
	denied, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	denied.Body.Close()
	if denied.StatusCode == 200 {
		t.Fatal("accepted traversal")
	}
	var readers sync.WaitGroup
	for i := 0; i < 4; i++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for j := 0; j < 5; j++ {
				response, err := http.Get(source)
				if err == nil {
					_, _ = io.Copy(io.Discard, response.Body)
					_ = response.Body.Close()
				}
			}
		}()
	}
	g.Release(source)
	readers.Wait()
	g.mu.RLock()
	remaining := len(g.urls)
	g.mu.RUnlock()
	if remaining != 0 {
		t.Fatalf("mux input resources remain: %d", remaining)
	}
	if _, err := os.Stat(session.dir); !os.IsNotExist(err) {
		t.Fatalf("temporary dir remains: %v", err)
	}
	resp, err = http.Get(source)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("released status: %s", resp.Status)
	}
}

func TestCancelMuxWhileWaitingForUpstream(t *testing.T) {
	if testing.Short() {
		t.Skip("ffmpeg integration test")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg unavailable")
	}
	started := make(chan struct{})
	var once sync.Once
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() { close(started) })
		<-r.Context().Done()
	}))
	defer upstream.Close()
	g, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() {
		_, err := g.RegisterMux(ctx, upstream.URL+"/video", "", upstream.URL+"/audio", false, 0)
		result <- err
	}()
	select {
	case <-started:
	case <-time.After(10 * time.Second):
		t.Fatal("ffmpeg did not request the input")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("mux did not stop after cancellation")
	}
	g.mu.RLock()
	resources, sessions := len(g.urls), len(g.muxes)
	g.mu.RUnlock()
	if resources != 0 || sessions != 0 {
		t.Fatalf("resources = %d, mux sessions = %d", resources, sessions)
	}
}
