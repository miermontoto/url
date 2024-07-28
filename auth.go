package main

import (
    "net/http"

    "github.com/dgrijalva/jwt-go"
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

func JWTAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        tokenString := r.Header.Get("Authorization")
        if tokenString == "" {
            http.Error(w, "Missing token", http.StatusUnauthorized)
            return
        }

        claims := &jwt.StandardClaims{}
        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
            return jwtKey, nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        r.Header.Set("X-User", claims.Subject)
        next.ServeHTTP(w, r)
    }
}
