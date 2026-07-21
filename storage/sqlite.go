package storage

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	db *sql.DB
}

var _ Storage = &SQLiteStorage{}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS urls (
            hash TEXT PRIMARY KEY,
            target TEXT NOT NULL,
            hits INTEGER DEFAULT 0,
            created DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated DATETIME DEFAULT CURRENT_TIMESTAMP,
            owner TEXT NOT NULL
        )
    `)
	if err != nil {
		return nil, err
	}

	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) Store(hash, target, owner string) error {
	_, err := s.db.Exec(
		"INSERT INTO urls (hash, target, owner) VALUES (?, ?, ?)",
		hash, target, owner)
	return err
}

func (s *SQLiteStorage) Get(hash string, ip string) (string, error) {
	var target string
	err := s.db.QueryRow(
		"SELECT target FROM urls WHERE hash = ?",
		hash).Scan(&target)
	if err == sql.ErrNoRows {
		return "", errors.New("URL not found")
	}
	if err != nil {
		return "", err
	}

	_, err = s.db.Exec(
		"UPDATE urls SET hits = hits + 1, updated = CURRENT_TIMESTAMP WHERE hash = ?",
		hash)

	// errors when updating hit count are not fatal.

	return target, nil
}

func (s *SQLiteStorage) GetURLInfo(hash string) (*URLInfo, error) {
	var info URLInfo
	err := s.db.QueryRow(
		"SELECT hash, target, hits, created, updated, owner FROM urls WHERE hash = ?",
		hash).Scan(&info.Hash, &info.Target, &info.Hits, &info.Created, &info.Updated, &info.Owner)
	if err == sql.ErrNoRows {
		return nil, errors.New("URL not found")
	}
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}


func (s *SQLiteStorage) Search(target string) ([]URLInfo, error) {
	rows, err := s.db.Query(
		"SELECT hash, target, hits, created, updated, owner FROM urls WHERE target LIKE ?",
		"%"+target+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []URLInfo
	for rows.Next() {
		var info URLInfo
		err := rows.Scan(&info.Hash, &info.Target, &info.Hits, &info.Created, &info.Updated, &info.Owner)
		if err != nil {
			return nil, err
		}
		results = append(results, info)
	}

	return results, nil
}

func (s *SQLiteStorage) SearchByOwner(owner string) ([]URLInfo, error) {
	rows, err := s.db.Query(
		"SELECT hash, target, hits, created, updated, owner FROM urls WHERE owner = ?",
		owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []URLInfo
	for rows.Next() {
		var info URLInfo
		err := rows.Scan(&info.Hash, &info.Target, &info.Hits, &info.Created, &info.Updated, &info.Owner)
		if err != nil {
			return nil, err
		}
		results = append(results, info)
	}

	return results, nil
}

func (s *SQLiteStorage) Delete(hash string) error {
	_, err := s.db.Exec("DELETE FROM urls WHERE hash = ?", hash)
	return err
}
