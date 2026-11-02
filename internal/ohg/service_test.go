package ohg

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/beilunyang/openhongguo/sdk"
)

// TestLiveSmoke 会访问真实服务端，只在手动执行时运行：
//
//	go test ./internal/ohg -run TestLiveSmoke -v
func TestLiveSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("live test")
	}
	svc, err := NewService(filepath.Join(t.TempDir(), "anonymous.json"))
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	page, err := svc.Discover(ctx, sdk.DiscoverOptions{})
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	t.Logf("discover: %d items, hasMore=%v", len(page.Items), page.HasMore)
	if len(page.Items) == 0 {
		t.Fatal("no items")
	}
	first := page.Items[0]
	t.Logf("first series: %s %q", first.ID, first.Title)

	eps, err := svc.ListEpisodes(ctx, first.ID, sdk.EpisodeOptions{})
	if err != nil {
		t.Fatalf("ListEpisodes: %v", err)
	}
	t.Logf("episodes: %d", len(eps.Items))
	if len(eps.Items) > 0 {
		t.Logf("ep1: %s %q", eps.Items[0].ID, eps.Items[0].Title)
	}
}
