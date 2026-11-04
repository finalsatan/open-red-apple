// Command webview-smoke runs the test-only WebView2 build with local fixtures.
package main

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// The package builder adds generated, untracked video fixtures before building.
//
//go:embed fixtures/*
var fixtures embed.FS

const controlAddress = "127.0.0.1:34827"

var mediaName = regexp.MustCompile(`^(video\.mp4|index\.m3u8|seg[0-9]+\.ts)$`)

type result struct {
	OK     bool     `json:"ok"`
	Checks []string `json:"checks"`
	Error  string   `json:"error"`
}

type controlServer struct {
	mu      sync.RWMutex
	harness string
	results chan result
	files   fs.FS
}

func (s *controlServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Private-Network", "true")
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	switch r.URL.Path {
	case "/config":
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", 405)
			return
		}
		s.mu.RLock()
		origin := s.harness
		s.mu.RUnlock()
		if origin == "" {
			http.Error(w, "not ready", 503)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"rpc": origin})
		return
	case "/report":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		var response result
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024)).Decode(&response); err != nil {
			http.Error(w, "invalid report", 400)
			return
		}
		select {
		case s.results <- response:
		default:
		}
		_, _ = io.WriteString(w, "ok")
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/")
	if !mediaName.MatchString(name) || (r.Method != http.MethodGet && r.Method != http.MethodHead) {
		http.NotFound(w, r)
		return
	}
	data, err := fs.ReadFile(s.files, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	switch {
	case strings.HasSuffix(name, ".m3u8"):
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	case strings.HasSuffix(name, ".ts"):
		w.Header().Set("Content-Type", "video/mp2t")
	default:
		w.Header().Set("Content-Type", "video/mp4")
	}
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
}

func replaceEnv(base []string, values map[string]string) []string {
	out := make([]string, 0, len(base)+len(values))
	for _, entry := range base {
		key, _, _ := strings.Cut(entry, "=")
		skip := false
		for candidate := range values {
			if strings.EqualFold(key, candidate) {
				skip = true
				break
			}
		}
		if !skip {
			out = append(out, entry)
		}
	}
	for key, value := range values {
		out = append(out, key+"="+value)
	}
	return out
}

func waitHarness(stdout io.Reader, done <-chan struct{}, processErr *error, timeout time.Duration) (string, error) {
	ready := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if value, ok := strings.CutPrefix(scanner.Text(), "WRA_HARNESS_URL="); ok {
				ready <- value
				return
			}
		}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case origin := <-ready:
		return origin, nil
	case <-done:
		return "", fmt.Errorf("测试后端提前退出：%v", *processErr)
	case <-timer.C:
		return "", errors.New("测试后端启动超时")
	}
}

func validateResult(r result) error {
	if !r.OK {
		if r.Error == "" {
			return errors.New("WebView2 自检失败")
		}
		return errors.New(r.Error)
	}
	have := make(map[string]bool, len(r.Checks))
	for _, check := range r.Checks {
		have[check] = true
	}
	for _, check := range []string{"mp4-resume", "mp4-seek", "sqlite-progress", "automatic-next", "hls-seek"} {
		if !have[check] {
			return fmt.Errorf("缺少自检结果：%s", check)
		}
	}
	return nil
}

func run() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	base := filepath.Dir(exe)
	appPath := filepath.Join(base, "win-red-apple-webview-test.exe")
	harnessPath := filepath.Join(base, "wra-player-harness.exe")
	for _, path := range []string{appPath, harnessPath} {
		if info, err := os.Stat(path); err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("缺少测试文件：%s", path)
		}
	}
	content, err := fs.Sub(fixtures, "fixtures/generated")
	if err != nil {
		return err
	}
	for _, name := range []string{"video.mp4", "index.m3u8", "seg00.ts"} {
		if _, err := fs.Stat(content, name); err != nil {
			return fmt.Errorf("测试素材不完整：%s", name)
		}
	}
	temp, err := os.MkdirTemp("", "win-red-apple-webview-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	appData := filepath.Join(temp, "appdata")
	localData := filepath.Join(temp, "localappdata")
	for _, path := range []string{appData, localData} {
		if err := os.MkdirAll(path, 0700); err != nil {
			return err
		}
	}
	listener, err := net.Listen("tcp4", controlAddress)
	if err != nil {
		return fmt.Errorf("本地测试端口不可用：%w", err)
	}
	control := &controlServer{files: content, results: make(chan result, 1)}
	server := &http.Server{Handler: control, ReadHeaderTimeout: 5 * time.Second}
	go server.Serve(listener)
	defer server.Close()
	harness := exec.Command(harnessPath, "-test.run=^TestPlayerHarness$", "-test.v")
	harness.Env = replaceEnv(os.Environ(), map[string]string{"WRA_E2E_MEDIA_ORIGIN": "http://" + controlAddress})
	stdout, err := harness.StdoutPipe()
	if err != nil {
		return err
	}
	harness.Stderr = os.Stderr
	if err = harness.Start(); err != nil {
		return fmt.Errorf("启动测试后端失败：%w", err)
	}
	harnessDone := make(chan struct{})
	var harnessErr error
	go func() { harnessErr = harness.Wait(); close(harnessDone) }()
	defer func() {
		if control.harness != "" {
			client := http.Client{Timeout: 2 * time.Second}
			response, e := client.Post(control.harness+"/shutdown", "text/plain", nil)
			if e == nil {
				response.Body.Close()
			}
		}
		select {
		case <-harnessDone:
		case <-time.After(3 * time.Second):
			_ = harness.Process.Kill()
			<-harnessDone
		}
	}()
	origin, err := waitHarness(stdout, harnessDone, &harnessErr, 20*time.Second)
	if err != nil {
		return err
	}
	control.mu.Lock()
	control.harness = origin
	control.mu.Unlock()
	app := exec.Command(appPath)
	app.Env = replaceEnv(os.Environ(), map[string]string{"AppData": appData, "LOCALAPPDATA": localData})
	if err = app.Start(); err != nil {
		return fmt.Errorf("启动测试程序失败：%w", err)
	}
	appDone := make(chan struct{})
	var appErr error
	go func() { appErr = app.Wait(); close(appDone) }()
	defer func() {
		select {
		case <-appDone:
		default:
			killProcessTree(app.Process)
			<-appDone
		}
	}()
	fmt.Println("测试窗口已启动，正在检查续播、拖动和自动下一集……")
	timer := time.NewTimer(90 * time.Second)
	defer timer.Stop()
	select {
	case r := <-control.results:
		if err := validateResult(r); err != nil {
			return err
		}
		fmt.Println("WebView2 自检通过：", strings.Join(r.Checks, ", "))
		return nil
	case <-appDone:
		return fmt.Errorf("测试程序在完成前退出：%v", appErr)
	case <-timer.C:
		return errors.New("WebView2 自检超时；请确认使用已登录的 Windows 桌面并安装 WebView2 Runtime")
	}
}

func main() {
	configureConsole()
	noPause := flag.Bool("no-pause", false, "exit immediately after the check")
	flag.Parse()
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "自检失败：", err)
	}
	if !*noPause {
		fmt.Print("按 Enter 键退出……")
		_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
	}
	if err != nil {
		os.Exit(1)
	}
}
