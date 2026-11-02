package store

import (
	"time"

	"github.com/finalsatan/open-red-apple/internal/dto"
)

// ToggleFavorite 切换收藏状态，返回切换后是否已收藏。
func (s *Store) ToggleFavorite(f dto.Favorite) (bool, error) {
	var exists int
	if err := s.db.QueryRow(
		`SELECT COUNT(1) FROM favorites WHERE series_id = ?`, f.SeriesID).Scan(&exists); err != nil {
		return false, err
	}
	if exists > 0 {
		_, err := s.db.Exec(`DELETE FROM favorites WHERE series_id = ?`, f.SeriesID)
		return false, err
	}
	addedAt := f.AddedAt
	if addedAt == 0 {
		addedAt = time.Now().Unix()
	}
	_, err := s.db.Exec(
		`INSERT INTO favorites(series_id, title, cover_url, added_at) VALUES(?, ?, ?, ?)`,
		f.SeriesID, f.Title, f.CoverURL, addedAt,
	)
	return true, err
}

// IsFavorite 报告一部短剧是否已收藏。
func (s *Store) IsFavorite(seriesID string) (bool, error) {
	var exists int
	if err := s.db.QueryRow(
		`SELECT COUNT(1) FROM favorites WHERE series_id = ?`, seriesID).Scan(&exists); err != nil {
		return false, err
	}
	return exists > 0, nil
}

// ListFavorites 返回收藏，按收藏时间倒序。
func (s *Store) ListFavorites() ([]dto.Favorite, error) {
	rows, err := s.db.Query(
		`SELECT series_id, title, cover_url, added_at FROM favorites ORDER BY added_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]dto.Favorite, 0)
	for rows.Next() {
		var f dto.Favorite
		if err := rows.Scan(&f.SeriesID, &f.Title, &f.CoverURL, &f.AddedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}
