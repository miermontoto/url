package storage

import "time"

type URLInfo struct {
	Hash    string    `json:"hash"`
	Target  string    `json:"target"`
	Hits    int       `json:"hits"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
	Owner   string    `json:"owner"`
}

type Storage interface {
	Store(hash, target, owner string) error
	Get(hash string) (string, error)
	Search(target string) ([]URLInfo, error)
	SearchByOwner(owner string) ([]URLInfo, error)
	GetURLInfo(hash string) (*URLInfo, error)
	CreateUser(username, password string) error
	AuthenticateUser(username, password string) bool
	Close() error
	Delete(hash string) error
}
