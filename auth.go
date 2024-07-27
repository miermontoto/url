package main

import (
    "net/http"
)

func DatabaseAuth(storage Storage) func(http.HandlerFunc) http.HandlerFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            username, password, ok := r.BasicAuth()
            if !ok || !storage.AuthenticateUser(username, password) {
                w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            r.Header.Set("X-User", username)
            next.ServeHTTP(w, r)
        }
    }
}
