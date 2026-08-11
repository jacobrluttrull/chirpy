package login_test

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
	"github.com/jacobrluttrull/chirpy/internal/login"
)

type fakeStore struct {
	user               database.User
	getUserErr         error
	createRefreshErr   error
	createdRefreshArgs database.CreateRefreshTokenParams
	createRefreshCalls int
}

func (f *fakeStore) GetUserEmail(ctx context.Context, email string) (database.User, error) {
	return f.user, f.getUserErr
}

func (f *fakeStore) CreateRefreshToken(ctx context.Context, arg database.CreateRefreshTokenParams) (database.RefreshToken, error) {
	f.createdRefreshArgs = arg
	f.createRefreshCalls++
	return database.RefreshToken{Token: arg.Token, UserID: arg.UserID, ExpiresAt: arg.ExpiresAt}, f.createRefreshErr
}

func TestHandlerLoginMalformedBody(t *testing.T) {
	cfg := &login.Config{}

	req := httptest.NewRequest("POST", "/api/login", strings.NewReader("not-json"))
	w := httptest.NewRecorder()

	cfg.HandlerLogin(w, req)

	if w.Code != 400 {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandlerLoginUnknownUser(t *testing.T) {
	cfg := &login.Config{DB: &fakeStore{getUserErr: errors.New("not found")}}

	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"email":"lane@example.com","password":"hunter2"}`))
	w := httptest.NewRecorder()

	cfg.HandlerLogin(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerLoginWrongPassword(t *testing.T) {
	hash, err := auth.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	cfg := &login.Config{DB: &fakeStore{user: database.User{HashedPassword: hash}}}

	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"email":"lane@example.com","password":"wrong-password"}`))
	w := httptest.NewRecorder()

	cfg.HandlerLogin(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerLoginSuccess(t *testing.T) {
	hash, err := auth.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	userID := uuid.New()
	store := &fakeStore{user: database.User{ID: userID, Email: "lane@example.com", HashedPassword: hash}}
	cfg := &login.Config{DB: store, JWTSecret: "test-secret"}

	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"email":"lane@example.com","password":"correct-password","expires_in_seconds":2}`))
	w := httptest.NewRecorder()

	cfg.HandlerLogin(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"token"`) {
		t.Fatalf("expected body to contain a token, got %q", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"refresh_token"`) {
		t.Fatalf("expected body to contain a refresh_token, got %q", w.Body.String())
	}
	if store.createRefreshCalls != 1 {
		t.Fatalf("expected CreateRefreshToken to be called once, got %d", store.createRefreshCalls)
	}
	if store.createdRefreshArgs.UserID != userID {
		t.Fatalf("expected refresh token to be created for user %v, got %v", userID, store.createdRefreshArgs.UserID)
	}
	wantExpiry := time.Now().Add(60 * 24 * time.Hour)
	if diff := wantExpiry.Sub(store.createdRefreshArgs.ExpiresAt); diff < -time.Minute || diff > time.Minute {
		t.Fatalf("expected refresh token to expire around %v, got %v", wantExpiry, store.createdRefreshArgs.ExpiresAt)
	}
}

func TestHandlerLoginRefreshTokenCreationFails(t *testing.T) {
	hash, err := auth.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	store := &fakeStore{
		user:             database.User{Email: "lane@example.com", HashedPassword: hash},
		createRefreshErr: errors.New("boom"),
	}
	cfg := &login.Config{DB: store, JWTSecret: "test-secret"}

	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"email":"lane@example.com","password":"correct-password"}`))
	w := httptest.NewRecorder()

	cfg.HandlerLogin(w, req)

	if w.Code != 500 {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
