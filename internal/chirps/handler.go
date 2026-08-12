package chirps

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/jacobrluttrull/chirpy/internal/auth"
	"github.com/jacobrluttrull/chirpy/internal/database"
	"github.com/jacobrluttrull/chirpy/internal/response"
)

type Store interface {
	CreateChirp(ctx context.Context, arg database.CreateChirpParams) (database.Chirp, error)
	GetAllChirps(ctx context.Context) ([]database.Chirp, error)
	GetChirp(ctx context.Context, id uuid.UUID) (database.Chirp, error)
	DeleteChirp(ctx context.Context, id uuid.UUID) error
}

type Config struct {
	DB        Store
	JWTSecret string
}

type returnValues struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func chirpToReturnValues(chirp database.Chirp) returnValues {
	return returnValues{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}
}

func (cfg *Config) HandlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	type parameters struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Something went wrong")
		return
	}
	if len(params.Body) > 140 {
		response.Error(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	cleanedBody := censorProfaneWords(params.Body)

	chirp, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userID,
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	response.JSON(w, http.StatusCreated, chirpToReturnValues(chirp))
}

func (cfg *Config) HandlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.DB.GetAllChirps(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	result := make([]returnValues, len(chirps))
	for i, chirp := range chirps {
		result[i] = chirpToReturnValues(chirp)
	}
	response.JSON(w, http.StatusOK, result)
}

func (cfg *Config) HandlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirp, err := cfg.DB.GetChirp(r.Context(), uuid.MustParse(r.PathValue("chirpID")))
	if errors.Is(err, sql.ErrNoRows) {
		response.Error(w, http.StatusNotFound, "Chirp not found")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	response.JSON(w, http.StatusOK, chirpToReturnValues(chirp))
}

func (cfg *Config) HandlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	chirp, err := cfg.DB.GetChirp(r.Context(), uuid.MustParse(r.PathValue("chirpID")))
	if errors.Is(err, sql.ErrNoRows) {
		response.Error(w, http.StatusNotFound, "Chirp not found")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	if chirp.UserID != userID {
		response.Error(w, http.StatusForbidden, "Forbidden")
		return
	}

	err = cfg.DB.DeleteChirp(r.Context(), chirp.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
