package users_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jacobrluttrull/chirpy/internal/database"
	"github.com/jacobrluttrull/chirpy/internal/users"
)

type fakeCreator struct {
	user database.User
	err  error
}

func (f *fakeCreator) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
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
	fake := &fakeCreator{user: database.User{Email: "lane@example.com"}}
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
	fake := &fakeCreator{err: errors.New("boom")}
	cfg := &users.Config{DB: fake}

	req := httptest.NewRequest("POST", "/api/users", strings.NewReader(`{"email":"lane@example.com","password":"hunter2"}`))
	w := httptest.NewRecorder()

	cfg.HandlerCreateUser(w, req)

	if w.Code != 500 {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
