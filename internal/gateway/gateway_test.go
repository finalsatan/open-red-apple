package gateway

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHLSFallbackWhenPlaylistBodyStalls(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		_, _ = w.Write([]byte("#EXTM3U\n"))
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer primary.Close()
	backup := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("#EXTM3U\n#EXTINF:4,\nsegment.ts\n"))
	}))
	defer backup.Close()
	g, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	g.playlistReadTimeout = 100 * time.Millisecond
	source, err := g.RegisterStream(primary.URL, backup.URL, true)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Timeout: 3 * time.Second}
	response, err := client.Get(source)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil || response.StatusCode != http.StatusOK || !strings.Contains(string(body), "#EXTINF:4,") {
		t.Fatalf("playlist %s, %q, %v", response.Status, body, err)
	}
}

func TestRangeAndAccess(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Range") != "bytes=2-3" {
			t.Errorf("range = %q", r.Header.Get("Range"))
		}
		w.Header().Set("Content-Range", "bytes 2-3/5")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write([]byte("cd"))
	}))
	defer upstream.Close()
	g, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	u, err := g.Register(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("Range", "bytes=2-3")
	req.Header.Set("Origin", "http://wails.localhost")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 206 || string(body) != "cd" || resp.Header.Get("Content-Range") != "bytes 2-3/5" {
		t.Fatalf("unexpected response: %s %q", resp.Status, body)
	}
	req, _ = http.NewRequest("GET", u, nil)
	req.Header.Set("Origin", "https://evil.example")
	denied, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer denied.Body.Close()
	if denied.StatusCode != http.StatusForbidden {
		t.Fatalf("unexpected status: %s", denied.Status)
	}
	req, _ = http.NewRequest("GET", strings.Replace(u, "127.0.0.1", "localhost", 1), nil)
	deniedHost, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer deniedHost.Body.Close()
	if deniedHost.StatusCode != http.StatusForbidden {
		t.Fatalf("unexpected host status: %s", deniedHost.Status)
	}
}

func TestHLSFallbackWhenPrimaryPlaylistIsUnusable(t *testing.T) {
	backup := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/backup/index.m3u8":
			_, _ = w.Write([]byte("#EXTM3U\n#EXTINF:4,\nsegment.ts\n"))
		case "/backup/segment.ts":
			_, _ = w.Write([]byte("backup segment"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer backup.Close()
	for _, tc := range []struct {
		name  string
		serve func(http.ResponseWriter)
	}{
		{"invalid content", func(w http.ResponseWriter) { _, _ = w.Write([]byte("not a playlist")) }},
		{"invalid URI", func(w http.ResponseWriter) { _, _ = w.Write([]byte("#EXTM3U\nhttps://[invalid\n")) }},
		{"truncated body", func(w http.ResponseWriter) {
			w.Header().Set("Content-Length", "100")
			_, _ = w.Write([]byte("#EXTM3U\n"))
		}},
		{"authentication error", func(w http.ResponseWriter) { http.Error(w, "unauthorized", http.StatusUnauthorized) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { tc.serve(w) }))
			defer primary.Close()
			g, err := New()
			if err != nil {
				t.Fatal(err)
			}
			defer g.Close()
			source, err := g.RegisterStream(primary.URL+"/index.m3u8", backup.URL+"/backup/index.m3u8", true)
			if err != nil {
				t.Fatal(err)
			}
			response, err := http.Get(source)
			if err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(response.Body)
			response.Body.Close()
			if err != nil || response.StatusCode != http.StatusOK || !strings.HasPrefix(string(data), "#EXTM3U\n") {
				t.Fatalf("playlist %s, %q, %v", response.Status, data, err)
			}
			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			segment, err := http.Get(lines[len(lines)-1])
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(segment.Body)
			segment.Body.Close()
			if err != nil || segment.StatusCode != http.StatusOK || string(body) != "backup segment" {
				t.Fatalf("segment %s, %q, %v", segment.Status, body, err)
			}
		})
	}
}

func TestMediaPreflight(t *testing.T) {
	g, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	source, err := g.Register("https://example.com/video.mp4")
	if err != nil {
		t.Fatal(err)
	}
	request := func(target, origin, method, headers string) *http.Response {
		t.Helper()
		req, err := http.NewRequest(http.MethodOptions, target, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Origin", origin)
		req.Header.Set("Access-Control-Request-Method", method)
		req.Header.Set("Access-Control-Request-Headers", headers)
		req.Header.Set("Access-Control-Request-Private-Network", "true")
		response, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		return response
	}
	response := request(source, "http://wails.localhost", "GET", "range")
	if response.StatusCode != http.StatusNoContent || response.Header.Get("Access-Control-Allow-Origin") != "http://wails.localhost" || response.Header.Get("Access-Control-Allow-Private-Network") != "true" {
		t.Fatalf("preflight: %s %v", response.Status, response.Header)
	}
	for _, test := range []struct {
		target, origin, method, headers string
		status                          int
	}{
		{source, "https://evil.example", "GET", "range", http.StatusForbidden},
		{source, "http://wails.localhost", "POST", "range", http.StatusForbidden},
		{source, "http://wails.localhost", "GET", "authorization", http.StatusForbidden},
		{source + "unknown", "http://wails.localhost", "GET", "range", http.StatusNotFound},
	} {
		if got := request(test.target, test.origin, test.method, test.headers).StatusCode; got != test.status {
			t.Fatalf("preflight = %d, want %d", got, test.status)
		}
	}
}

func TestHLSRewriteAndBackup(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/broken":
			http.Error(w, "unavailable", 503)
		case "/dir/master.m3u8":
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			_, _ = w.Write([]byte("#EXTM3U\n#EXT-X-MAP:URI=\"init.mp4\"\nchild.m3u8?x=1\n"))
		case "/dir/child.m3u8":
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			_, _ = w.Write([]byte("#EXTM3U\nsegment.ts\n"))
		case "/dir/segment.ts":
			_, _ = w.Write([]byte("segment"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()
	g, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	u, err := g.RegisterStream(upstream.URL+"/broken", upstream.URL+"/dir/master.m3u8", true)
	if err != nil {
		t.Fatal(err)
	}
	get := func(u string) string {
		t.Helper()
		resp, err := http.Get(u)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("status %s", resp.Status)
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	master := get(u)
	lines := strings.Split(strings.TrimSpace(master), "\n")
	if len(lines) != 3 || !strings.Contains(lines[1], "URI=\"http://127.0.0.1:") || !strings.HasSuffix(lines[2], "/index.m3u8") {
		t.Fatalf("master: %s", master)
	}
	child := get(lines[2])
	childLines := strings.Split(strings.TrimSpace(child), "\n")
	if len(childLines) != 2 || get(childLines[1]) != "segment" {
		t.Fatalf("child: %s", child)
	}
	second, err := g.RegisterStream(upstream.URL+"/broken", upstream.URL+"/dir/master.m3u8", true)
	if err != nil {
		t.Fatal(err)
	}
	if second == u {
		t.Fatal("independent playbacks share a token")
	}
	g.Release(u)
	for _, released := range []string{u, lines[2], childLines[1]} {
		response, err := http.Get(released)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusNotFound {
			t.Fatalf("released %s: %s", released, response.Status)
		}
	}
	if !strings.Contains(get(second), "#EXTM3U") {
		t.Fatal("release affected another session")
	}
}

func TestHLSPlaylistWithoutExtension(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/master":
			_, _ = w.Write([]byte("#EXTM3U\n#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"a\",URI=\"audio\"\n#EXT-X-STREAM-INF:BANDWIDTH=1000\nvideo?token=1\n"))
		case "/video", "/audio":
			_, _ = w.Write([]byte("#EXTM3U\nsegment.ts\n"))
		case "/segment.ts":
			_, _ = w.Write([]byte("segment"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()
	g, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	u, err := g.RegisterStream(upstream.URL+"/master", "", true)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get(u)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	if len(lines) != 4 || !strings.Contains(lines[1], "/index.m3u8\"") || !strings.HasSuffix(lines[3], "/index.m3u8") {
		t.Fatalf("master: %s", body)
	}
	child, err := http.Get(lines[3])
	if err != nil {
		t.Fatal(err)
	}
	defer child.Body.Close()
	childBody, _ := io.ReadAll(child.Body)
	if !strings.Contains(string(childBody), "http://127.0.0.1:") {
		t.Fatalf("child: %s", childBody)
	}
}

func TestHLSFallbackForNestedResources(t *testing.T) {
	backup := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/backup/child.m3u8":
			_, _ = w.Write([]byte("#EXTM3U\n#EXTINF:4,\nsegment.ts\n"))
		case "/backup/segment.ts":
			_, _ = w.Write([]byte("backup segment"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer backup.Close()
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/primary/master.m3u8":
			_, _ = w.Write([]byte("#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=1000\nchild.m3u8\n"))
		case "/primary/child.m3u8", "/primary/segment.ts":
			http.Error(w, "CDN unavailable", http.StatusServiceUnavailable)
		default:
			http.NotFound(w, r)
		}
	}))
	defer primary.Close()
	g, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	source, err := g.RegisterStream(primary.URL+"/primary/master.m3u8", backup.URL+"/backup/master.m3u8", true)
	if err != nil {
		t.Fatal(err)
	}
	get := func(raw string) string {
		t.Helper()
		response, err := http.Get(raw)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("status: %s", response.Status)
		}
		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		return string(body)
	}
	masterLines := strings.Split(strings.TrimSpace(get(source)), "\n")
	childLines := strings.Split(strings.TrimSpace(get(masterLines[len(masterLines)-1])), "\n")
	if segment := get(childLines[len(childLines)-1]); segment != "backup segment" {
		t.Fatalf("segment: %q", segment)
	}
}

func TestRejectInvalidStreamURL(t *testing.T) {
	g, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	for _, raw := range []string{"", "file:///tmp/video", "http://user:pass@example.com/video"} {
		if _, err := g.RegisterStream(raw, "", false); err == nil {
			t.Errorf("accepted %q", raw)
		}
	}
}
