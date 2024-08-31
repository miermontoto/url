package main

import (
	"log"
	"net/http"
	"os"
	"github.com/miermontoto/url/storage"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func loadEnv(key string) string {
	godotenv.Load()
	return os.Getenv(key)
}

var jwtKey = []byte(loadEnv("JWT_KEY"))

func main() {
	// storage, err := NewSQLiteStorage("urls.db")
	storage, err := storage.NewPostgresStorage(loadEnv("PSQL_CONN"))

	if err != nil {
		log.Fatal(err)
	}
	defer storage.Close()

	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Serve index.html
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	}).Methods("GET")

	r.HandleFunc("/auth", AuthHandler(storage)).Methods("POST")

	r.HandleFunc("/search", JWTAuth((SearchHandler(storage)))).Methods("GET")
	r.HandleFunc("/shorten", JWTAuth((ShortenHandler(storage)))).Methods("POST")
	r.HandleFunc("/info/{hash}", JWTAuth((URLInfoHandler(storage)))).Methods("GET")
	r.HandleFunc("/my", JWTAuth((MyURLsHandler(storage)))).Methods("GET")
	r.HandleFunc("/{hash}", RedirectHandler(storage)).Methods("GET")

	port := loadEnv("PORT")
	if port == "" {
		port = "4343"
	}

	log.Println("server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
