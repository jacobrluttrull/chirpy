package admin

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/jacobrluttrull/chirpy/internal/response"
)

type UserDeleter interface {
	DeleteAllUsers(ctx context.Context) error
}

type Config struct {
	DB       UserDeleter
	Platform string

	hits atomic.Int32
}

func (cfg *Config) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.hits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *Config) FileserverHitsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`<html>
        <body>
          <h1>Welcome, Chirpy Admin</h1>
          <p>Chirpy has been visited %d times!</p>
        </body>
      </html>`, cfg.hits.Load())))
}

func (cfg *Config) ResetFileserverHitsHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		response.Error(w, http.StatusForbidden, "Forbidden")
		return
	}
	err := cfg.DB.DeleteAllUsers(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete users")
		return
	}
	cfg.hits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits counter reset"))
}
