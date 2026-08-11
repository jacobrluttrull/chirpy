package refresh_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/jacobrluttrull/chirpy/internal/database"
	"github.com/jacobrluttrull/chirpy/internal/refresh"
)

type fakeStore struct {
	user         database.User
	getUserErr   error
	revokeErr    error
	revokedToken string
	revokeCalled bool
}

func (f *fakeStore) GetUserFromRefreshToken(ctx context.Context, token string) (database.User, error) {
	return f.user, f.getUserErr
}

func (f *fakeStore) RevokeRefreshToken(ctx context.Context, token string) error {
	f.revokeCalled = true
	f.revokedToken = token
	return f.revokeErr
}

func TestHandlerRefreshNoAuthHeader(t *testing.T) {
	cfg := &refresh.Config{DB: &fakeStore{}, JWTSecret: "test-secret"}

	req := httptest.NewRequest("POST", "/api/refresh", nil)
	w := httptest.NewRecorder()

	cfg.HandlerRefresh(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerRefreshUnknownOrExpiredToken(t *testing.T) {
	cfg := &refresh.Config{DB: &fakeStore{getUserErr: sql.ErrNoRows}, JWTSecret: "test-secret"}

	req := httptest.NewRequest("POST", "/api/refresh", nil)
	req.Header.Set("Authorization", "Bearer some-refresh-token")
	w := httptest.NewRecorder()

	cfg.HandlerRefresh(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerRefreshSuccess(t *testing.T) {
	userID := uuid.New()
	cfg := &refresh.Config{DB: &fakeStore{user: database.User{ID: userID}}, JWTSecret: "test-secret"}

	req := httptest.NewRequest("POST", "/api/refresh", nil)
	req.Header.Set("Authorization", "Bearer some-refresh-token")
	w := httptest.NewRecorder()

	cfg.HandlerRefresh(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"token"`) {
		t.Fatalf("expected body to contain a token, got %q", w.Body.String())
	}
}

func TestHandlerRevokeNoAuthHeader(t *testing.T) {
	cfg := &refresh.Config{DB: &fakeStore{}}

	req := httptest.NewRequest("POST", "/api/revoke", nil)
	w := httptest.NewRecorder()

	cfg.HandlerRevoke(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerRevokeSuccess(t *testing.T) {
	store := &fakeStore{}
	cfg := &refresh.Config{DB: store}

	req := httptest.NewRequest("POST", "/api/revoke", nil)
	req.Header.Set("Authorization", "Bearer some-refresh-token")
	w := httptest.NewRecorder()

	cfg.HandlerRevoke(w, req)

	if w.Code != 204 {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
	if !store.revokeCalled {
		t.Fatal("expected RevokeRefreshToken to be called")
	}
	if store.revokedToken != "some-refresh-token" {
		t.Fatalf("expected token %q to be revoked, got %q", "some-refresh-token", store.revokedToken)
	}
}

func TestHandlerRevokeDBError(t *testing.T) {
	cfg := &refresh.Config{DB: &fakeStore{revokeErr: errors.New("boom")}}

	req := httptest.NewRequest("POST", "/api/revoke", nil)
	req.Header.Set("Authorization", "Bearer some-refresh-token")
	w := httptest.NewRecorder()

	cfg.HandlerRevoke(w, req)

	if w.Code != 500 {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
