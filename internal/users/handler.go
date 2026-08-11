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

type Creator interface {
	CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error)
}

type Config struct {
	DB Creator
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
	type returnValues struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}
	response.JSON(w, http.StatusCreated, returnValues{
		ID:        user.ID.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}
