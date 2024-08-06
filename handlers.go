package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

type JSendSuccess struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type JSendError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type ExistingURL struct {
	URL     string `json:"url"`
	Hash    string `json:"hash"`
	Existed bool   `json:"existed"`
}

func buildRedirectUrl(r *http.Request, hash string) string {
	return r.Host + "/" + hash
}

func successResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(JSendSuccess{Status: "success", Data: data})
}

func failureResponse(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(JSendError{Status: "error", Message: message})
}

func AuthHandler(storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !storage.AuthenticateUser(creds.Username, creds.Password) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		expirationTime := time.Now().Add(24 * time.Hour)
		claims := &jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   creds.Username,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenResponse{Token: tokenString})
	}
}

func ShortenHandler(storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			failureResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		hash := r.URL.Query().Get("hash")

		// check if the URL is already in the database
		info, err2 := storage.Search(req.URL)

		// if it is, return the short URL
		if len(info) != 0 && err2 == nil && hash == "" {
			exists := ExistingURL{
				URL:     buildRedirectUrl(r, info[0].Hash),
				Hash:    info[0].Hash,
				Existed: true,
			}

			successResponse(w, exists)
			return
		}

		if hash != "" {
			// check if the hash is already in the database
			_, err := storage.Get(hash)
			if err == nil {
				failureResponse(w, "Hash already exists", http.StatusConflict)
				return
			}
		}

		if hash == "" {
			hash = generateHash(storage)
		}
		owner := r.Header.Get("X-User")
		err := storage.Store(hash, req.URL, owner)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		exists := ExistingURL{
			URL:     buildRedirectUrl(r, hash),
			Hash:    hash,
			Existed: false,
		}

		successResponse(w, exists)
	}
}

func RedirectHandler(storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := mux.Vars(r)["hash"]

		target, err := storage.Get(hash)
		if err != nil {
			failureResponse(w, "URL not found", http.StatusNotFound)
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
			failureResponse(w, "URL not found", http.StatusNotFound)
			return
		}

		successResponse(w, info)
	}
}

func SearchHandler(storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			failureResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		results, err := storage.Search(req.URL)
		if err != nil {
			failureResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		successResponse(w, results)
	}
}

func MyURLsHandler(storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner := r.Header.Get("X-User")
		results, err := storage.SearchByOwner(owner)
		if err != nil {
			failureResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		successResponse(w, results)
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
