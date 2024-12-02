package storage

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type PostgresStorage struct {
	db *sql.DB
}

var _ Storage = &PostgresStorage{}

func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			hash TEXT PRIMARY KEY,
			target TEXT NOT NULL,
			created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			owner TEXT NOT NULL
		)
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			username TEXT PRIMARY KEY,
			password TEXT NOT NULL
		)
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS target_idx ON urls (target)
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS hits (
			hash TEXT PRIMARY KEY REFERENCES urls(hash),
			hits INTEGER DEFAULT 0,
			ip TEXT NOT NULL,
			time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) Store(hash, target, owner string) error {
	_, err := s.db.Exec(
		"INSERT INTO urls (hash, target, owner) VALUES ($1, $2, $3)",
		hash, target, owner)
	return err
}

func (s *PostgresStorage) Get(hash string, ip string) (string, error) {
	var target string
	err := s.db.QueryRow(
		"SELECT target FROM urls WHERE hash = $1",
		hash).Scan(&target)
	if err == sql.ErrNoRows {
		return "", errors.New("URL not found")
	}
	if err != nil {
		return "", err
	}

	if ip == "api" {
		return target, nil
	}

	_, err = s.db.Exec(
		"INSERT INTO hits (hash, ip) VALUES ($1, '$2') ON CONFLICT (hash) DO UPDATE SET hits = hits + 1",
		hash, ip)

	// errors when updating hit count are not fatal.
	return target, nil
}

func (s *PostgresStorage) GetURLInfo(hash string) (*URLInfo, error) {
	var info URLInfo
	err := s.db.QueryRow(
		"SELECT hash, target, hits, created, updated, owner FROM urls WHERE hash = $1",
		hash).Scan(&info.Hash, &info.Target, &info.Hits, &info.Created, &info.Updated, &info.Owner)
	if err == sql.ErrNoRows {
		return nil, errors.New("URL not found")
	}
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

func (s *PostgresStorage) CreateUser(username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", username, string(hashedPassword))
	return err
}

func (s *PostgresStorage) AuthenticateUser(username, password string) bool {
	var storedPassword string
	err := s.db.QueryRow("SELECT password FROM users WHERE username = $1", username).Scan(&storedPassword)
	if err != nil {
		return false
	}
	return password == storedPassword
}

func (s *PostgresStorage) Search(target string) ([]URLInfo, error) {
	rows, err := s.db.Query(
		"SELECT hash, target, hits, created, updated, owner FROM urls WHERE target LIKE $1",
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

func (s *PostgresStorage) SearchByOwner(owner string) ([]URLInfo, error) {
	rows, err := s.db.Query(
		"SELECT hash, target, hits, created, updated, owner FROM urls WHERE owner = $1",
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

func (s *PostgresStorage) Delete(hash string) error {
	_, err := s.db.Exec("DELETE FROM urls WHERE hash = $1", hash)
	return err
}
