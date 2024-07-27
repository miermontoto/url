package main

import (
    "log"
    "net/http"
	"os"

    "github.com/gorilla/mux"
)

func main() {
    storage, err := NewSQLiteStorage("urls.db")
    if err != nil {
        log.Fatal(err)
    }
    defer storage.Close()

	r := mux.NewRouter()

	r.HandleFunc("/search", DatabaseAuth(storage)(SearchHandler(storage))).Methods("GET")
    r.HandleFunc("/shorten", DatabaseAuth(storage)(ShortenHandler(storage))).Methods("POST")
    r.HandleFunc("/info/{hash}", DatabaseAuth(storage)(URLInfoHandler(storage))).Methods("GET")
	r.HandleFunc("/{hash}", RedirectHandler(storage)).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {port = "4343"}

    log.Println("server running on port", port)
    log.Fatal(http.ListenAndServe(":"+port, r))
}
