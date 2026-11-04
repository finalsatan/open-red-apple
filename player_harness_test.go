package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/beilunyang/openhongguo/sdk"
	"github.com/beilunyang/openhongguo/sdk/model"
	"github.com/finalsatan/open-red-apple/internal/dto"
	"github.com/finalsatan/open-red-apple/internal/gateway"
	"github.com/finalsatan/open-red-apple/internal/ohg"
	"github.com/finalsatan/open-red-apple/internal/store"
)

// TestPlayerHarness is a local-only bridge for the browser integration test.
// It uses actual App methods, the media gateway and SQLite, but supplies a
// deterministic playback model instead of contacting the remote service.
func TestPlayerHarness(t *testing.T) {
	mediaOrigin := os.Getenv("WRA_E2E_MEDIA_ORIGIN")
	if mediaOrigin == "" {
		t.Skip("browser integration harness")
	}
	st, err := store.Open(filepath.Join(t.TempDir(), "history.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	g, err := gateway.New()
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	var slowStarted atomic.Bool
	a := &App{svc: &ohg.Service{}, store: st, media: g, playClient: fakePlaybackClient{
		resolve: func(ctx context.Context, id string, _ sdk.PlaybackOptions) (*model.Playback, error) {
			if id == "slow" {
				slowStarted.Store(true)
				<-ctx.Done()
				return nil, ctx.Err()
			}
			if id == "ep2" {
				return &model.Playback{Streams: []model.Stream{{URL: mediaOrigin + "/index.m3u8", HLS: true, Codec: "h264"}}}, nil
			}
			if id == "ep1" {
				return &model.Playback{Streams: []model.Stream{{URL: mediaOrigin + "/video.mp4", Codec: "h264"}}}, nil
			}
			return nil, fmt.Errorf("unknown episode")
		}, reset: func(context.Context) error { return nil },
	}}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	var server *http.Server
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = server.Shutdown(ctx)
		}()
	})
	mux.HandleFunc("/rpc", func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:34115" || origin == "http://wails.localhost" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions && (origin == "http://localhost:34115" || origin == "http://wails.localhost") {
			w.Header().Set("Access-Control-Allow-Methods", "POST")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Private-Network", "true")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var call struct {
			Method    string          `json:"method"`
			ID        string          `json:"id"`
			RequestID string          `json:"requestId"`
			URL       string          `json:"url"`
			Item      dto.HistoryItem `json:"item"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024)).Decode(&call); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		var value any
		var err error
		switch call.Method {
		case "IsSlowStarted":
			value = slowStarted.Load()
		case "BeginPlayback":
			value, err = a.BeginPlayback()
		case "CancelPlayback":
			a.CancelPlayback(call.RequestID)
		case "PlayURLForRequest":
			value, err = a.PlayURLForRequest(call.ID, call.RequestID)
		case "PlayURL":
			value, err = a.PlayURL(call.ID)
		case "GetProgress":
			value, err = a.GetProgress(call.ID)
		case "SaveProgress":
			err = a.SaveProgress(call.Item)
		case "ReleasePlayback":
			a.ReleasePlayback(call.URL)
		default:
			http.Error(w, "unknown method", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		_ = json.NewEncoder(w).Encode(value)
	})
	server = &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	fmt.Printf("WRA_HARNESS_URL=http://%s\n", listener.Addr())
	if err = server.Serve(listener); err != nil && err != http.ErrServerClosed {
		t.Fatal(err)
	}
}
