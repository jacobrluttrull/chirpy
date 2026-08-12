package users

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jacobrluttrull/chirpy/internal/auth"
	"github.com/jacobrluttrull/chirpy/internal/database"
	"github.com/jacobrluttrull/chirpy/internal/response"
)

type Store interface {
	CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error)
	UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error)
}

type Config struct {
	DB        Store
	JWTSecret string
}

type userResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func toUserResponse(user database.User) userResponse {
	return userResponse{
		ID:        user.ID.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
}

func (cfg *Config) HandlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	params := request{}
	err := decoder.Decode(&params)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Something went wrong")
		return
	}
	if params.Password == "" {
		response.Error(w, http.StatusBadRequest, "Password is required")
		return
	}
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create user")
		return
	}
	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create user")
		return
	}
	response.JSON(w, http.StatusCreated, toUserResponse(user))
}

func (cfg *Config) HandlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Something went wrong")
		return
	}
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update user")
		return
	}
	user, err := cfg.DB.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update user")
		return
	}
	response.JSON(w, http.StatusOK, toUserResponse(user))
}
