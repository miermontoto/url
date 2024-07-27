package main

import (
    "encoding/json"
    "net/http"
	"math/rand"

    "github.com/gorilla/mux"
)

type ShortenRequest struct {
    URL string `json:"url"`
}

type ShortenResponse struct {
    ShortURL string `json:"url"`
}

func ShortenHandler(storage Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req ShortenRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        hash := generateHash(storage)
        owner := r.Header.Get("X-User")
        err := storage.Store(hash, req.URL, owner)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        resp := ShortenResponse{ShortURL: r.Host + "/" + hash}
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(resp)
    }
}

func RedirectHandler(storage Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        hash := mux.Vars(r)["hash"]
        target, err := storage.Get(hash)
        if err != nil {
            http.Error(w, "URL not found", http.StatusNotFound)
            return
        }
        http.Redirect(w, r, target, http.StatusFound)
    }
}

func URLInfoHandler(storage Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        hash := mux.Vars(r)["hash"]
        info, err := storage.GetURLInfo(hash)
        if err != nil {
            http.Error(w, "URL not found", http.StatusNotFound)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(info)
    }
}

func generateHash(storage Storage) string {
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    const length = 3

    for {
        hash := make([]byte, length)
        for i := range hash {
            hash[i] = charset[rand.Intn(len(charset))]
        }

        _, err := storage.Get(string(hash))
        if err != nil {
            return string(hash)
        }
    }
}
