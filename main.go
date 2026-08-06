package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"time"

	"github.com/jacobrluttrull/chirpy/internal/database"
)
type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform      string
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{Error: msg})
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) fileserverHitsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`<html>
        <body>
          <h1>Welcome, Chirpy Admin</h1>
          <p>Chirpy has been visited %d times!</p>
        </body>
      </html>`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) resetFileserverHitsHandler(w http.ResponseWriter, r *http.Request) {
    if cfg.platform != "dev" {
        respondWithError(w, 403, "Forbidden")
        return
    }
    err := cfg.db.DeleteAllUsers(r.Context())
    if err != nil {
        respondWithError(w, 500, "Failed to delete users")
        return
    }
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits counter reset"))
}

var profaneWords = map[string]bool{
	"kerfuffle": true,
	"sharbert":  true,
	"fornax":    true,
}

func censorProfaneWords(body string) string {
	words := strings.Fields(body)
	for i, word := range words {
		if profaneWords[strings.ToLower(word)] {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type Chirp struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := Chirp{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Something went wrong")
		return
	}
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	cleanedBody := censorProfaneWords(params.Body)
	type returnValues struct {
		CleanedBody string `json:"cleaned_body,omitempty"`
	}
	respondWithJSON(w, 200, returnValues{CleanedBody: cleanedBody})
}

func (cfg *apiConfig) handlerCreateUser (w http.ResponseWriter, r *http.Request) {
    type request struct {
        Email string `json:"email"`
    }
    decoder := json.NewDecoder(r.Body)
    params := request{}
    err := decoder.Decode(&params)
    if err != nil {
        respondWithError(w, 400, "Something went wrong")
        return
    }
    user, err := cfg.db.CreateUser(r.Context(), params.Email)
    if err != nil {
        respondWithError(w, 500, "Failed to create user")
        return
    }
    type returnValues struct {
        ID        string    `json:"id"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`
        Email     string    `json:"email"`
    }
    respondWithJSON(w, 201, returnValues{
        ID:        user.ID.String(),
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
        Email:     user.Email,
    })
}

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	dbQueries := database.New(db)

	platform := os.Getenv("PLATFORM")

	mux := http.NewServeMux()
	apiCfg := &apiConfig{
		db: dbQueries,
		platform: platform,
	}
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.Handle("/api/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("GET /admin/metrics", apiCfg.fileserverHitsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetFileserverHitsHandler)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	srv.ListenAndServe()
}
