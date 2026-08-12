package polka

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jacobrluttrull/chirpy/internal/auth"
	"github.com/jacobrluttrull/chirpy/internal/database"
)

type Store interface {
	UpgradeUserToChirpyRed(ctx context.Context, id uuid.UUID) (database.User, error)
}

type Config struct {
	DB     Store
	APIKey string
}

func (cfg *Config) HandlerChirpyRedWebhooks(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetApiKey(r.Header)
	if err != nil || apiKey != cfg.APIKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	type request struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}
	decoder := json.NewDecoder(r.Body)
	params := request{}
	err = decoder.Decode(&params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	userID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err = cfg.DB.UpgradeUserToChirpyRed(r.Context(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
