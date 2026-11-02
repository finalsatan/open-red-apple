package store

import "time"

// AddSearch 记录一个搜索词，重复搜索会刷新时间。
func (s *Store) AddSearch(keyword string) error {
	_, err := s.db.Exec(
		`INSERT INTO search_history(keyword, updated_at) VALUES(?, ?)
		 ON CONFLICT(keyword) DO UPDATE SET updated_at=excluded.updated_at`,
		keyword, time.Now().Unix(),
	)
	return err
}

// SearchHistory 返回最近的搜索词。
func (s *Store) SearchHistory(limit int) ([]string, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.Query(
		`SELECT keyword FROM search_history ORDER BY updated_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]string, 0, limit)
	for rows.Next() {
		var kw string
		if err := rows.Scan(&kw); err != nil {
			return nil, err
		}
		out = append(out, kw)
	}
	return out, rows.Err()
}

// ClearSearchHistory 清空搜索记录。
func (s *Store) ClearSearchHistory() error {
	_, err := s.db.Exec(`DELETE FROM search_history`)
	return err
}
