package db

import (
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
)

func Open(path string) (*sql.DB, error) {
  return sql.Open("sqlite3", path)
}

type Store struct {
  DB *sql.DB
}

func NewStore(db *sql.DB) *Store {
  return &Store{DB: db}
}
