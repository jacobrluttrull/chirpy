package polka_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/jacobrluttrull/chirpy/internal/database"
	"github.com/jacobrluttrull/chirpy/internal/polka"
)

type fakeStore struct {
	upgradedUser database.User
	upgradeErr   error
	called       bool
	calledID     uuid.UUID
}

func (f *fakeStore) UpgradeUserToChirpyRed(ctx context.Context, id uuid.UUID) (database.User, error) {
	f.called = true
	f.calledID = id
	return f.upgradedUser, f.upgradeErr
}

func TestHandlerChirpyRedWebhooksNoAuthHeader(t *testing.T) {
	cfg := &polka.Config{DB: &fakeStore{}, APIKey: "test-key"}

	req := httptest.NewRequest("POST", "/api/polka/webhooks", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	cfg.HandlerChirpyRedWebhooks(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerChirpyRedWebhooksWrongAPIKey(t *testing.T) {
	cfg := &polka.Config{DB: &fakeStore{}, APIKey: "right-key"}

	req := httptest.NewRequest("POST", "/api/polka/webhooks", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "ApiKey wrong-key")
	w := httptest.NewRecorder()

	cfg.HandlerChirpyRedWebhooks(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerChirpyRedWebhooksMalformedBody(t *testing.T) {
	cfg := &polka.Config{DB: &fakeStore{}, APIKey: "test-key"}

	req := httptest.NewRequest("POST", "/api/polka/webhooks", strings.NewReader(`not-json`))
	req.Header.Set("Authorization", "ApiKey test-key")
	w := httptest.NewRecorder()

	cfg.HandlerChirpyRedWebhooks(w, req)

	if w.Code != 400 {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandlerChirpyRedWebhooksIgnoresOtherEvents(t *testing.T) {
	store := &fakeStore{}
	cfg := &polka.Config{DB: store, APIKey: "test-key"}

	req := httptest.NewRequest("POST", "/api/polka/webhooks", strings.NewReader(`{"event":"user.downgraded","data":{"user_id":"`+uuid.New().String()+`"}}`))
	req.Header.Set("Authorization", "ApiKey test-key")
	w := httptest.NewRecorder()

	cfg.HandlerChirpyRedWebhooks(w, req)

	if w.Code != 204 {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
	if store.called {
		t.Fatal("expected UpgradeUserToChirpyRed not to be called for a non-upgrade event")
	}
}

func TestHandlerChirpyRedWebhooksInvalidUserID(t *testing.T) {
	cfg := &polka.Config{DB: &fakeStore{}, APIKey: "test-key"}

	req := httptest.NewRequest("POST", "/api/polka/webhooks", strings.NewReader(`{"event":"user.upgraded","data":{"user_id":"not-a-uuid"}}`))
	req.Header.Set("Authorization", "ApiKey test-key")
	w := httptest.NewRecorder()

	cfg.HandlerChirpyRedWebhooks(w, req)

	if w.Code != 400 {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandlerChirpyRedWebhooksUserNotFound(t *testing.T) {
	store := &fakeStore{upgradeErr: sql.ErrNoRows}
	cfg := &polka.Config{DB: store, APIKey: "test-key"}

	req := httptest.NewRequest("POST", "/api/polka/webhooks", strings.NewReader(`{"event":"user.upgraded","data":{"user_id":"`+uuid.New().String()+`"}}`))
	req.Header.Set("Authorization", "ApiKey test-key")
	w := httptest.NewRecorder()

	cfg.HandlerChirpyRedWebhooks(w, req)

	if w.Code != 404 {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestHandlerChirpyRedWebhooksDBError(t *testing.T) {
	store := &fakeStore{upgradeErr: errors.New("boom")}
	cfg := &polka.Config{DB: store, APIKey: "test-key"}

	req := httptest.NewRequest("POST", "/api/polka/webhooks", strings.NewReader(`{"event":"user.upgraded","data":{"user_id":"`+uuid.New().String()+`"}}`))
	req.Header.Set("Authorization", "ApiKey test-key")
	w := httptest.NewRecorder()

	cfg.HandlerChirpyRedWebhooks(w, req)

	if w.Code != 500 {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestHandlerChirpyRedWebhooksSuccess(t *testing.T) {
	userID := uuid.New()
	store := &fakeStore{upgradedUser: database.User{ID: userID, IsChirpyRed: true}}
	cfg := &polka.Config{DB: store, APIKey: "test-key"}

	req := httptest.NewRequest("POST", "/api/polka/webhooks", strings.NewReader(`{"event":"user.upgraded","data":{"user_id":"`+userID.String()+`"}}`))
	req.Header.Set("Authorization", "ApiKey test-key")
	w := httptest.NewRecorder()

	cfg.HandlerChirpyRedWebhooks(w, req)

	if w.Code != 204 {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
	if !store.called {
		t.Fatal("expected UpgradeUserToChirpyRed to be called")
	}
	if store.calledID != userID {
		t.Fatalf("expected UpgradeUserToChirpyRed to be called with %v, got %v", userID, store.calledID)
	}
}
