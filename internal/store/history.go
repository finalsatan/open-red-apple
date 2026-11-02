package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/finalsatan/open-red-apple/internal/dto"
)

// SaveProgress 写入或更新一部短剧的观看进度。
func (s *Store) SaveProgress(h dto.HistoryItem) error {
	updatedAt := h.UpdatedAt
	if updatedAt == 0 {
		updatedAt = time.Now().Unix()
	}
	finished := 0
	if h.Finished {
		finished = 1
	}
	_, err := s.db.Exec(
		`INSERT INTO watch_history
		   (series_id, episode_id, episode_number, title, cover_url, position, duration, finished, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(series_id) DO UPDATE SET
		   episode_id=excluded.episode_id,
		   episode_number=excluded.episode_number,
		   title=excluded.title,
		   cover_url=excluded.cover_url,
		   position=excluded.position,
		   duration=excluded.duration,
		   finished=excluded.finished,
		   updated_at=excluded.updated_at`,
		h.SeriesID, h.EpisodeID, h.EpisodeNumber, h.Title, h.CoverURL,
		h.Position, h.Duration, finished, updatedAt,
	)
	return err
}

// GetProgress 返回一部短剧的观看进度；没有记录时返回 nil。
func (s *Store) GetProgress(seriesID string) (*dto.HistoryItem, error) {
	row := s.db.QueryRow(
		`SELECT series_id, episode_id, episode_number, title, cover_url, position, duration, finished, updated_at
		 FROM watch_history WHERE series_id = ?`, seriesID)
	h, err := scanHistory(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// ListHistory 返回观看记录，按最近观看倒序。
func (s *Store) ListHistory(limit int) ([]dto.HistoryItem, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(
		`SELECT series_id, episode_id, episode_number, title, cover_url, position, duration, finished, updated_at
		 FROM watch_history ORDER BY updated_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]dto.HistoryItem, 0)
	for rows.Next() {
		h, err := scanHistory(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// DeleteHistory 删除一部短剧的观看记录。
func (s *Store) DeleteHistory(seriesID string) error {
	_, err := s.db.Exec(`DELETE FROM watch_history WHERE series_id = ?`, seriesID)
	return err
}

// ClearHistory 清空观看记录。
func (s *Store) ClearHistory() error {
	_, err := s.db.Exec(`DELETE FROM watch_history`)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanHistory(sc scanner) (dto.HistoryItem, error) {
	var h dto.HistoryItem
	var finished int
	err := sc.Scan(&h.SeriesID, &h.EpisodeID, &h.EpisodeNumber, &h.Title, &h.CoverURL,
		&h.Position, &h.Duration, &finished, &h.UpdatedAt)
	h.Finished = finished != 0
	return h, err
}
