package refresh

import (
	"context"
	"net/http"
	"time"

	"github.com/jacobrluttrull/chirpy/internal/auth"
	"github.com/jacobrluttrull/chirpy/internal/database"
	"github.com/jacobrluttrull/chirpy/internal/response"
)

const accessTokenExpiration = time.Hour

type Store interface {
	GetUserFromRefreshToken(ctx context.Context, token string) (database.User, error)
	RevokeRefreshToken(ctx context.Context, token string) error
}

type Config struct {
	DB        Store
	JWTSecret string
}

func (cfg *Config) HandlerRefresh(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := cfg.DB.GetUserFromRefreshToken(r.Context(), tokenString)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.JWTSecret, accessTokenExpiration)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create token")
		return
	}

	type returnValues struct {
		Token string `json:"token"`
	}
	response.JSON(w, http.StatusOK, returnValues{Token: accessToken})
}

func (cfg *Config) HandlerRevoke(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err = cfg.DB.RevokeRefreshToken(r.Context(), tokenString)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to revoke token")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
