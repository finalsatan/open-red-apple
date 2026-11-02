package store

import (
	"path/filepath"
	"testing"

	"github.com/finalsatan/open-red-apple/internal/dto"
)

func TestStoreRoundTrip(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	// 搜索记录：去重且最近的在前。
	for _, kw := range []string{"末世", "玄幻", "末世"} {
		if err := s.AddSearch(kw); err != nil {
			t.Fatalf("AddSearch(%q): %v", kw, err)
		}
	}
	history, err := s.SearchHistory(10)
	if err != nil {
		t.Fatalf("SearchHistory: %v", err)
	}
	if len(history) != 2 || history[0] != "末世" {
		t.Fatalf("SearchHistory = %v, want [末世 玄幻]", history)
	}

	// 收藏：切换两次应回到未收藏。
	fav := dto.Favorite{SeriesID: "s1", Title: "剧一", CoverURL: "c"}
	on, err := s.ToggleFavorite(fav)
	if err != nil || !on {
		t.Fatalf("ToggleFavorite on = %v, %v", on, err)
	}
	if list, _ := s.ListFavorites(); len(list) != 1 || list[0].SeriesID != "s1" {
		t.Fatalf("ListFavorites = %v", list)
	}
	if on, _ := s.ToggleFavorite(fav); on {
		t.Fatal("ToggleFavorite should turn the favorite off")
	}
	if list, _ := s.ListFavorites(); len(list) != 0 {
		t.Fatalf("ListFavorites after off = %v", list)
	}

	// 观看记录：同一部剧只保留最近一条。
	if err := s.SaveProgress(dto.HistoryItem{
		SeriesID: "s1", EpisodeID: "e1", EpisodeNumber: 1,
		Title: "剧一", CoverURL: "c", Position: 30, Duration: 120,
	}); err != nil {
		t.Fatalf("SaveProgress: %v", err)
	}
	if err := s.SaveProgress(dto.HistoryItem{
		SeriesID: "s1", EpisodeID: "e2", EpisodeNumber: 2,
		Title: "剧一", CoverURL: "c", Position: 5, Duration: 120,
	}); err != nil {
		t.Fatalf("SaveProgress: %v", err)
	}
	got, err := s.GetProgress("s1")
	if err != nil || got == nil {
		t.Fatalf("GetProgress = %v, %v", got, err)
	}
	if got.EpisodeNumber != 2 || got.EpisodeID != "e2" {
		t.Fatalf("GetProgress = %+v, want episode 2", got)
	}
	if list, _ := s.ListHistory(10); len(list) != 1 {
		t.Fatalf("ListHistory = %v, want 1 row", list)
	}
	if err := s.ClearHistory(); err != nil {
		t.Fatalf("ClearHistory: %v", err)
	}
	if got, _ := s.GetProgress("s1"); got != nil {
		t.Fatalf("GetProgress after clear = %+v", got)
	}
}
