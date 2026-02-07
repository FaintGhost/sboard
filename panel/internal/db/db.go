package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

func Open(path string) (*sql.DB, error) {
	return sql.Open("sqlite3", path)
}

type Store struct {
	DB  *sql.DB
	Now func() time.Time
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		DB:  db,
		Now: time.Now,
	}
}

func (s *Store) NowUTC() time.Time {
	if s == nil {
		return time.Now().UTC()
	}
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}
