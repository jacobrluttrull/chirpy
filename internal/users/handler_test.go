package users_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jacobrluttrull/chirpy/internal/auth"
	"github.com/jacobrluttrull/chirpy/internal/database"
	"github.com/jacobrluttrull/chirpy/internal/users"
)

type fakeStore struct {
	user database.User
	err  error
}

func (f *fakeStore) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	return f.user, f.err
}

func (f *fakeStore) UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error) {
	return f.user, f.err
}

func TestHandlerCreateUserMalformedBody(t *testing.T) {
	cfg := &users.Config{}

	req := httptest.NewRequest("POST", "/api/users", strings.NewReader("not-json"))
	w := httptest.NewRecorder()

	cfg.HandlerCreateUser(w, req)

	if w.Code != 400 {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandlerCreateUserMissingPassword(t *testing.T) {
	cfg := &users.Config{}

	req := httptest.NewRequest("POST", "/api/users", strings.NewReader(`{"email":"lane@example.com"}`))
	w := httptest.NewRecorder()

	cfg.HandlerCreateUser(w, req)

	if w.Code != 400 {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandlerCreateUserSuccess(t *testing.T) {
	fake := &fakeStore{user: database.User{Email: "lane@example.com"}}
	cfg := &users.Config{DB: fake}

	req := httptest.NewRequest("POST", "/api/users", strings.NewReader(`{"email":"lane@example.com","password":"hunter2"}`))
	w := httptest.NewRecorder()

	cfg.HandlerCreateUser(w, req)

	if w.Code != 201 {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "lane@example.com") {
		t.Fatalf("expected body to contain the created email, got %q", w.Body.String())
	}
}

func TestHandlerCreateUserDBError(t *testing.T) {
	fake := &fakeStore{err: errors.New("boom")}
	cfg := &users.Config{DB: fake}

	req := httptest.NewRequest("POST", "/api/users", strings.NewReader(`{"email":"lane@example.com","password":"hunter2"}`))
	w := httptest.NewRecorder()

	cfg.HandlerCreateUser(w, req)

	if w.Code != 500 {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestHandlerUpdateUserMissingToken(t *testing.T) {
	cfg := &users.Config{JWTSecret: "test-secret"}

	req := httptest.NewRequest("PUT", "/api/users", strings.NewReader(`{"email":"new@example.com","password":"hunter2"}`))
	w := httptest.NewRecorder()

	cfg.HandlerUpdateUser(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerUpdateUserMalformedToken(t *testing.T) {
	cfg := &users.Config{JWTSecret: "test-secret"}

	req := httptest.NewRequest("PUT", "/api/users", strings.NewReader(`{"email":"new@example.com","password":"hunter2"}`))
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	w := httptest.NewRecorder()

	cfg.HandlerUpdateUser(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerUpdateUserSuccess(t *testing.T) {
	userID := uuid.New()
	token, err := auth.MakeJWT(userID, "test-secret", time.Hour)
	if err != nil {
		t.Fatalf("failed to make token: %v", err)
	}
	fake := &fakeStore{user: database.User{ID: userID, Email: "new@example.com"}}
	cfg := &users.Config{DB: fake, JWTSecret: "test-secret"}

	req := httptest.NewRequest("PUT", "/api/users", strings.NewReader(`{"email":"new@example.com","password":"hunter2"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	cfg.HandlerUpdateUser(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "new@example.com") {
		t.Fatalf("expected body to contain the updated email, got %q", w.Body.String())
	}
	if strings.Contains(w.Body.String(), "hunter2") || strings.Contains(w.Body.String(), "password") {
		t.Fatalf("expected body to omit the password, got %q", w.Body.String())
	}
}
