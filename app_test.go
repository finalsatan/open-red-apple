package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/beilunyang/openhongguo/sdk"
	"github.com/beilunyang/openhongguo/sdk/model"
	"github.com/finalsatan/open-red-apple/internal/gateway"
	"github.com/finalsatan/open-red-apple/internal/ohg"
)

type fakePlaybackClient struct {
	resolve func(context.Context, string, sdk.PlaybackOptions) (*model.Playback, error)
	reset   func(context.Context) error
}

func (f fakePlaybackClient) ResolvePlayback(ctx context.Context, episodeID string, options sdk.PlaybackOptions) (*model.Playback, error) {
	return f.resolve(ctx, episodeID, options)
}
func (f fakePlaybackClient) ResetDevice(ctx context.Context) error { return f.reset(ctx) }

func TestPlaybackRetriesAndServesBackup(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/primary" {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		if r.Header.Get("Range") != "bytes=2-3" {
			t.Errorf("missing range: %q", r.Header.Get("Range"))
		}
		w.Header().Set("Content-Range", "bytes 2-3/5")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write([]byte("cd"))
	}))
	defer upstream.Close()
	g, err := gateway.New()
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	var calls, resets int
	client := fakePlaybackClient{
		resolve: func(_ context.Context, id string, _ sdk.PlaybackOptions) (*model.Playback, error) {
			calls++
			if id != "episode" {
				t.Errorf("episode = %q", id)
			}
			if calls == 1 {
				return nil, sdk.ErrAnonymousLimited
			}
			return &model.Playback{Streams: []model.Stream{{URL: upstream.URL + "/primary", BackupURL: upstream.URL + "/backup", Codec: "h264"}}}, nil
		},
		reset: func(context.Context) error { resets++; return nil },
	}
	a := &App{svc: &ohg.Service{}, media: g, playClient: client}
	source, err := a.PlayURL("episode")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || resets != 1 {
		t.Fatalf("calls = %d, resets = %d", calls, resets)
	}
	req, _ := http.NewRequest(http.MethodGet, source, nil)
	req.Header.Set("Range", "bytes=2-3")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil || resp.StatusCode != http.StatusPartialContent || string(body) != "cd" {
		t.Fatalf("response = %s %q, %v", resp.Status, body, err)
	}
	a.ReleasePlayback(source)
	resp, err = http.Get(source)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("released resource: %s", resp.Status)
	}
}

func TestPlaybackReportsDeviceLimitAndProtectedMedia(t *testing.T) {
	for _, tc := range []struct {
		name             string
		result           error
		want             string
		attempts, resets int
	}{
		{"limit", sdk.ErrAnonymousLimited, "匿名设备", 3, 2},
		{"protected", sdk.ErrDRMProtected, "加密", 1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls, resets := 0, 0
			a := &App{svc: &ohg.Service{}, playClient: fakePlaybackClient{
				resolve: func(context.Context, string, sdk.PlaybackOptions) (*model.Playback, error) {
					calls++
					return nil, tc.result
				},
				reset: func(context.Context) error { resets++; return nil },
			}}
			_, err := a.PlayURL("episode")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v", err)
			}
			if tc.name == "limit" && !errors.Is(err, sdk.ErrAnonymousLimited) {
				t.Fatalf("lost underlying error: %v", err)
			}
			if calls != tc.attempts || resets != tc.resets {
				t.Fatalf("calls = %d, resets = %d", calls, resets)
			}
		})
	}
}

func TestPlaybackResolutionSerializesDeviceUse(t *testing.T) {
	var active, maxActive atomic.Int32
	a := &App{playClient: fakePlaybackClient{
		resolve: func(context.Context, string, sdk.PlaybackOptions) (*model.Playback, error) {
			current := active.Add(1)
			for previous := maxActive.Load(); current > previous && !maxActive.CompareAndSwap(previous, current); previous = maxActive.Load() {
			}
			time.Sleep(10 * time.Millisecond)
			active.Add(-1)
			return &model.Playback{}, nil
		},
		reset: func(context.Context) error { return nil },
	}}
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := a.resolvePlayback(context.Background(), "episode"); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if got := maxActive.Load(); got != 1 {
		t.Fatalf("concurrent use = %d", got)
	}
}

func TestPlaybackCancellationReleasesDeviceAndCancelsQueuedRequest(t *testing.T) {
	g, err := gateway.New()
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	started := make(chan struct{})
	var calls atomic.Int32
	a := &App{svc: &ohg.Service{}, media: g, playClient: fakePlaybackClient{
		resolve: func(ctx context.Context, id string, _ sdk.PlaybackOptions) (*model.Playback, error) {
			calls.Add(1)
			if id == "slow" {
				close(started)
				<-ctx.Done()
				// The SDK may wrap a canceled transport as its transport error.
				return nil, sdk.ErrTransport
			}
			return &model.Playback{Streams: []model.Stream{{URL: "https://example.com/clear.mp4"}}}, nil
		},
		reset: func(context.Context) error { return nil },
	}}
	first, err := a.BeginPlayback()
	if err != nil {
		t.Fatal(err)
	}
	firstResult := make(chan error, 1)
	go func() { _, err := a.PlayURLForRequest("slow", first); firstResult <- err }()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("first request did not start")
	}
	second, err := a.BeginPlayback()
	if err != nil {
		t.Fatal(err)
	}
	secondResult := make(chan error, 1)
	go func() { _, err := a.PlayURLForRequest("queued", second); secondResult <- err }()
	a.CancelPlayback(second)
	select {
	case err := <-secondResult:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("queued request: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("queued request did not cancel while device was busy")
	}
	a.CancelPlayback(first)
	select {
	case err := <-firstResult:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("running request: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("running request did not cancel")
	}
	third, err := a.BeginPlayback()
	if err != nil {
		t.Fatal(err)
	}
	source, err := a.PlayURLForRequest("ready", third)
	if err != nil {
		t.Fatal(err)
	}
	a.ReleasePlayback(source)
	if calls.Load() != 2 {
		t.Fatalf("resolver calls = %d; canceled queued request used the device", calls.Load())
	}
	if _, err := a.PlayURLForRequest("ready", third); !errors.Is(err, context.Canceled) {
		t.Fatalf("reused request: %v", err)
	}
}

func TestSelectClearStream(t *testing.T) {
	p := &model.Playback{Streams: []model.Stream{
		{URL: "https://example.com/hevc", Codec: "bytevc1", Height: 1080},
		{URL: "https://example.com/h264", Codec: "h264", Height: 720},
		{URL: "https://example.com/protected", Codec: "h264", Height: 2160, Protected: true},
		{URL: "https://example.com/dash", Codec: "h264", Height: 1440, Dash: true},
	}}
	s, err := selectClearStream(p)
	if err != nil || s.URL != "https://example.com/h264" {
		t.Fatalf("selected %+v, %v", s, err)
	}
	p.Streams = p.Streams[:1]
	s, err = selectClearStream(p)
	if err != nil || s.Codec != "bytevc1" {
		t.Fatalf("fallback %+v, %v", s, err)
	}
	p.Streams = []model.Stream{{URL: "https://example.com/dash", Dash: true}}
	if _, err = selectClearStream(p); err == nil {
		t.Fatal("accepted unsupported stream")
	}
	p.AudioURL = "https://example.com/audio"
	if s, err = selectClearStream(p); err != nil || !s.Dash {
		t.Fatalf("did not select video with separate audio: %+v, %v", s, err)
	}
}
