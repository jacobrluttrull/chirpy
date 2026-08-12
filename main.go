package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/jacobrluttrull/chirpy/internal/admin"
	"github.com/jacobrluttrull/chirpy/internal/chirps"
	"github.com/jacobrluttrull/chirpy/internal/database"
	"github.com/jacobrluttrull/chirpy/internal/login"
	"github.com/jacobrluttrull/chirpy/internal/refresh"
	"github.com/jacobrluttrull/chirpy/internal/users"
)

func main() {
	godotenv.Load()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable not set")
	}

	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	dbQueries := database.New(db)

	adminCfg := &admin.Config{DB: dbQueries, Platform: os.Getenv("PLATFORM")}
	usersCfg := &users.Config{DB: dbQueries, JWTSecret: jwtSecret}
	loginCfg := &login.Config{DB: dbQueries, JWTSecret: jwtSecret}
	refreshCfg := &refresh.Config{DB: dbQueries, JWTSecret: jwtSecret}
	chirpsCfg := &chirps.Config{DB: dbQueries, JWTSecret: jwtSecret}

	mux := http.NewServeMux()
	mux.Handle("/app/", adminCfg.MiddlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.Handle("/api/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("GET /admin/metrics", adminCfg.FileserverHitsHandler)
	mux.HandleFunc("POST /admin/reset", adminCfg.ResetFileserverHitsHandler)
	mux.HandleFunc("POST /api/users", usersCfg.HandlerCreateUser)
	mux.HandleFunc("PUT /api/users", usersCfg.HandlerUpdateUser)
	mux.HandleFunc("POST /api/login", loginCfg.HandlerLogin)
	mux.HandleFunc("POST /api/refresh", refreshCfg.HandlerRefresh)
	mux.HandleFunc("POST /api/revoke", refreshCfg.HandlerRevoke)
	mux.HandleFunc("POST /api/chirps", chirpsCfg.HandlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", chirpsCfg.HandlerGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", chirpsCfg.HandlerGetChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", chirpsCfg.HandlerDeleteChirp)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	srv.ListenAndServe()
}
