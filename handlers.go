package main

import (
    "encoding/json"
    "net/http"
	"math/rand"
	"log"

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

		// check if the URL is already in the database
		info, err2 := storage.Search(req.URL)
		if err2 == nil {
			// if it is, return the short URL
			resp := ShortenResponse{ShortURL: r.Host + "/" + info[0].Hash}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
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

func SearchHandler(storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		results, err := storage.Search(req.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
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
