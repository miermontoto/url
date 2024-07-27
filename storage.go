package main

import "time"

type URLInfo struct {
    Hash    string
    Target  string
    Hits    int
    Created time.Time
    Updated time.Time
    Owner   string
}

type Storage interface {
    Store(hash, target, owner string) error
    Get(hash string) (string, error)
    GetURLInfo(hash string) (*URLInfo, error)
    CreateUser(username, password string) error
    AuthenticateUser(username, password string) bool
    Close() error
}
