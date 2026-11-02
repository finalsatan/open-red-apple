// Package store 用 SQLite 保存搜索记录、收藏和观看记录。
package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Store 是一个 SQLite 数据库。
type Store struct {
	db *sql.DB
}

// Open 打开（必要时创建）数据库并执行建表。
func Open(path string) (*Store, error) {
	dsn := path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// Close 关闭数据库。
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS search_history (
  keyword    TEXT PRIMARY KEY,
  updated_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS favorites (
  series_id TEXT PRIMARY KEY,
  title     TEXT NOT NULL,
  cover_url TEXT NOT NULL,
  added_at  INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS watch_history (
  series_id      TEXT PRIMARY KEY,
  episode_id     TEXT NOT NULL,
  episode_number INTEGER NOT NULL,
  title          TEXT NOT NULL,
  cover_url      TEXT NOT NULL,
  position       REAL NOT NULL,
  duration       REAL NOT NULL,
  finished       INTEGER NOT NULL,
  updated_at     INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_watch_updated ON watch_history(updated_at DESC);`
	_, err := s.db.Exec(schema)
	return err
}
