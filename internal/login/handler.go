package login

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jacobrluttrull/chirpy/internal/auth"
	"github.com/jacobrluttrull/chirpy/internal/database"
	"github.com/jacobrluttrull/chirpy/internal/response"
)

const (
	maxExpiration          = time.Hour
	refreshTokenExpiration = 60 * 24 * time.Hour
)

type Store interface {
	GetUserEmail(ctx context.Context, email string) (database.User, error)
	CreateRefreshToken(ctx context.Context, arg database.CreateRefreshTokenParams) (database.RefreshToken, error)
}

type Config struct {
	DB        Store
	JWTSecret string
}

// resolveExpiration clamps a client-requested token lifetime (in seconds)
// to maxExpiration. A non-positive value means "not specified" and falls
// back to the default of maxExpiration.
func resolveExpiration(requestedSeconds int) time.Duration {
	if requestedSeconds <= 0 {
		return maxExpiration
	}
	requested := time.Duration(requestedSeconds) * time.Second
	if requested > maxExpiration {
		return maxExpiration
	}
	return requested
}

func (cfg *Config) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	decoder := json.NewDecoder(r.Body)
	params := request{}
	err := decoder.Decode(&params)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Something went wrong")
		return
	}
	user, err := cfg.DB.GetUserEmail(r.Context(), params.Email)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}
	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	if !match {
		response.Error(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.JWTSecret, resolveExpiration(params.ExpiresInSeconds))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create token")
		return
	}

	refreshTokenString, err := auth.MakeRefreshToken()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create token")
		return
	}
	_, err = cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshTokenString,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(refreshTokenExpiration),
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create token")
		return
	}

	type returnValues struct {
		ID           string    `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}
	response.JSON(w, http.StatusOK, returnValues{
		ID:           user.ID.String(),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshTokenString,
	})
}
