package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/miermontoto/url/storage"
)

func loadEnv(key string) string {
	godotenv.Load()
	return os.Getenv(key)
}

func main() {
	// storage, err := NewSQLiteStorage("urls.db")
	storage, err := storage.NewPostgresStorage(loadEnv("PSQL_CONN"))

	if err != nil {
		log.Fatal(err)
	}
	defer storage.Close()

	// cliente oidc de mier id (backend-for-frontend): login/callback/logout y el
	// middleware RequireAuth que protege las rutas privadas
	a := newAuth(oidcConfig{
		issuer:       loadEnv("OIDC_ISSUER"),
		clientID:     loadEnv("OIDC_CLIENT_ID"),
		clientSecret: loadEnv("OIDC_CLIENT_SECRET"),
		redirectURL:  loadEnv("OIDC_REDIRECT_URL"),
		postLogout:   loadEnv("OIDC_POST_LOGOUT_URL"),
	})

	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Serve index.html
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	}).Methods("GET")

	// flujo oidc
	r.HandleFunc("/auth/login", a.LoginHandler).Methods("GET")
	r.HandleFunc("/auth/callback", a.CallbackHandler).Methods("GET")
	r.HandleFunc("/auth/logout", a.LogoutHandler).Methods("GET")
	r.HandleFunc("/auth/me", a.RequireAuth(a.MeHandler)).Methods("GET")

	// rutas privadas protegidas por la sesión oidc
	r.HandleFunc("/search", a.RequireAuth(SearchHandler(storage))).Methods("GET")
	r.HandleFunc("/shorten", a.RequireAuth(ShortenHandler(storage))).Methods("POST")
	r.HandleFunc("/info/{hash}", a.RequireAuth(URLInfoHandler(storage))).Methods("GET")
	r.HandleFunc("/my", a.RequireAuth(MyURLsHandler(storage))).Methods("GET")
	r.HandleFunc("/{hash}", RedirectHandler(storage)).Methods("GET")
	r.HandleFunc("/my/{hash}", a.RequireAuth(DeleteURLHandler(storage))).Methods("DELETE")

	port := loadEnv("PORT")
	if port == "" {
		port = "4343"
	}

	log.Println("server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
