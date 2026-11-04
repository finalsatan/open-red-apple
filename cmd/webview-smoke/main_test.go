package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestControlServer(t *testing.T) {
	control := &controlServer{files: fstest.MapFS{"video.mp4": &fstest.MapFile{Data: []byte("abcdef")}}, results: make(chan result, 1)}
	server := httptest.NewServer(control)
	defer server.Close()
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/video.mp4", nil)
	req.Header.Set("Range", "bytes=2-3")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusPartialContent || string(body) != "cd" || resp.Header.Get("Content-Range") != "bytes 2-3/6" {
		t.Fatalf("range: %s %q", resp.Status, body)
	}
	control.mu.Lock()
	control.harness = "http://127.0.0.1:1234"
	control.mu.Unlock()
	resp, err = http.Get(server.URL + "/config")
	if err != nil {
		t.Fatal(err)
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(body), "127.0.0.1:1234") {
		t.Fatalf("config: %s", body)
	}
	resp, err = http.Post(server.URL+"/report", "text/plain", strings.NewReader(`{"ok":true,"checks":["mp4-resume","mp4-seek","sqlite-progress","automatic-next","hls-seek"]}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("report status: %s", resp.Status)
	}
	select {
	case result := <-control.results:
		if err := validateResult(result); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatal("missing report")
	}
	if err := validateResult(result{OK: true, Checks: []string{"mp4-resume"}}); err == nil {
		t.Fatal("accepted incomplete report")
	}
}

func TestReplaceEnvIsCaseInsensitive(t *testing.T) {
	got := replaceEnv([]string{"AppData=old", "APPDATA=older", "OTHER=value"}, map[string]string{"APPDATA": "new"})
	joined := strings.Join(got, "|")
	if strings.Contains(joined, "old") || !strings.Contains(joined, "APPDATA=new") || !strings.Contains(joined, "OTHER=value") {
		t.Fatalf("environment: %v", got)
	}
}
