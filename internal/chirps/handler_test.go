package chirps_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/jacobrluttrull/chirpy/internal/chirps"
	"github.com/jacobrluttrull/chirpy/internal/database"
)

type fakeStore struct {
	createChirp    database.Chirp
	createErr      error
	allChirps      []database.Chirp
	allErr         error
	byAuthorChirps []database.Chirp
	byAuthorErr    error
	byAuthorID     uuid.UUID
	singleChirp    database.Chirp
	singleErr      error
	createCalled   bool
	createdParams  database.CreateChirpParams
	deleteErr      error
	deleteCalled   bool
	deletedID      uuid.UUID
}

func (f *fakeStore) CreateChirp(ctx context.Context, arg database.CreateChirpParams) (database.Chirp, error) {
	f.createCalled = true
	f.createdParams = arg
	return f.createChirp, f.createErr
}

func (f *fakeStore) GetAllChirps(ctx context.Context) ([]database.Chirp, error) {
	return f.allChirps, f.allErr
}

func (f *fakeStore) GetChirpsByAuthor(ctx context.Context, userID uuid.UUID) ([]database.Chirp, error) {
	f.byAuthorID = userID
	return f.byAuthorChirps, f.byAuthorErr
}

func (f *fakeStore) GetChirp(ctx context.Context, id uuid.UUID) (database.Chirp, error) {
	return f.singleChirp, f.singleErr
}

func (f *fakeStore) DeleteChirp(ctx context.Context, id uuid.UUID) error {
	f.deleteCalled = true
	f.deletedID = id
	return f.deleteErr
}

// signTestToken builds a JWT the same way auth.MakeJWT does, so the handler
// can be exercised without depending on auth package internals.
func signTestToken(t *testing.T, userID uuid.UUID, secret string, expiresIn time.Duration) string {
	t.Helper()

	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}
	return signed
}

func TestHandlerCreateChirpNoAuthHeader(t *testing.T) {
	cfg := &chirps.Config{DB: &fakeStore{}, JWTSecret: "test-secret"}

	req := httptest.NewRequest("POST", "/api/chirps", strings.NewReader(`{"body":"hello"}`))
	w := httptest.NewRecorder()

	cfg.HandlerCreateChirp(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerCreateChirpMalformedAuthHeader(t *testing.T) {
	cfg := &chirps.Config{DB: &fakeStore{}, JWTSecret: "test-secret"}

	req := httptest.NewRequest("POST", "/api/chirps", strings.NewReader(`{"body":"hello"}`))
	req.Header.Set("Authorization", "not-a-bearer-header")
	w := httptest.NewRecorder()

	cfg.HandlerCreateChirp(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerCreateChirpInvalidToken(t *testing.T) {
	cfg := &chirps.Config{DB: &fakeStore{}, JWTSecret: "test-secret"}

	req := httptest.NewRequest("POST", "/api/chirps", strings.NewReader(`{"body":"hello"}`))
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	w := httptest.NewRecorder()

	cfg.HandlerCreateChirp(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerCreateChirpWrongSecret(t *testing.T) {
	cfg := &chirps.Config{DB: &fakeStore{}, JWTSecret: "right-secret"}
	token := signTestToken(t, uuid.New(), "wrong-secret", time.Hour)

	req := httptest.NewRequest("POST", "/api/chirps", strings.NewReader(`{"body":"hello"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	cfg.HandlerCreateChirp(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerCreateChirpExpiredToken(t *testing.T) {
	cfg := &chirps.Config{DB: &fakeStore{}, JWTSecret: "test-secret"}
	token := signTestToken(t, uuid.New(), "test-secret", -time.Hour)

	req := httptest.NewRequest("POST", "/api/chirps", strings.NewReader(`{"body":"hello"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	cfg.HandlerCreateChirp(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerCreateChirpTooLong(t *testing.T) {
	cfg := &chirps.Config{DB: &fakeStore{}, JWTSecret: "test-secret"}
	token := signTestToken(t, uuid.New(), "test-secret", time.Hour)

	longBody := strings.Repeat("a", 141)
	req := httptest.NewRequest("POST", "/api/chirps", strings.NewReader(`{"body":"`+longBody+`"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	cfg.HandlerCreateChirp(w, req)

	if w.Code != 400 {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandlerCreateChirpSuccessUsesTokenUserID(t *testing.T) {
	userID := uuid.New()
	store := &fakeStore{createChirp: database.Chirp{ID: uuid.New(), UserID: userID, Body: "hello"}}
	cfg := &chirps.Config{DB: store, JWTSecret: "test-secret"}
	token := signTestToken(t, userID, "test-secret", time.Hour)

	req := httptest.NewRequest("POST", "/api/chirps", strings.NewReader(`{"body":"hello kerfuffle"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	cfg.HandlerCreateChirp(w, req)

	if w.Code != 201 {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
	if !store.createCalled {
		t.Fatal("expected CreateChirp to be called")
	}
	if store.createdParams.UserID != userID {
		t.Fatalf("expected chirp to be created for token's user %v, got %v", userID, store.createdParams.UserID)
	}
	if store.createdParams.Body != "hello ****" {
		t.Fatalf("expected profanity to be censored, got %q", store.createdParams.Body)
	}
}

func TestHandlerGetAllChirps(t *testing.T) {
	store := &fakeStore{allChirps: []database.Chirp{{Body: "one"}, {Body: "two"}}}
	cfg := &chirps.Config{DB: store}

	req := httptest.NewRequest("GET", "/api/chirps", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetAllChirps(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "one") || !strings.Contains(w.Body.String(), "two") {
		t.Fatalf("expected body to contain both chirps, got %q", w.Body.String())
	}
}

func TestHandlerGetAllChirpsSortDesc(t *testing.T) {
	now := time.Now()
	store := &fakeStore{allChirps: []database.Chirp{
		{Body: "oldest", CreatedAt: now},
		{Body: "newest", CreatedAt: now.Add(time.Hour)},
	}}
	cfg := &chirps.Config{DB: store}

	req := httptest.NewRequest("GET", "/api/chirps?sort=desc", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetAllChirps(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	body := w.Body.String()
	if strings.Index(body, "newest") > strings.Index(body, "oldest") {
		t.Fatalf("expected newest chirp before oldest in descending order, got %q", body)
	}
}

func TestHandlerGetAllChirpsSortAscDefault(t *testing.T) {
	now := time.Now()
	store := &fakeStore{allChirps: []database.Chirp{
		{Body: "oldest", CreatedAt: now},
		{Body: "newest", CreatedAt: now.Add(time.Hour)},
	}}
	cfg := &chirps.Config{DB: store}

	req := httptest.NewRequest("GET", "/api/chirps", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetAllChirps(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	body := w.Body.String()
	if strings.Index(body, "oldest") > strings.Index(body, "newest") {
		t.Fatalf("expected oldest chirp before newest in ascending order, got %q", body)
	}
}

func TestHandlerGetAllChirpsFilteredByAuthor(t *testing.T) {
	authorID := uuid.New()
	store := &fakeStore{
		allChirps:      []database.Chirp{{Body: "one"}, {Body: "two"}},
		byAuthorChirps: []database.Chirp{{Body: "mine", UserID: authorID}},
	}
	cfg := &chirps.Config{DB: store}

	req := httptest.NewRequest("GET", "/api/chirps?author_id="+authorID.String(), nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetAllChirps(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if store.byAuthorID != authorID {
		t.Fatalf("expected GetChirpsByAuthor to be called with %v, got %v", authorID, store.byAuthorID)
	}
	if !strings.Contains(w.Body.String(), "mine") || strings.Contains(w.Body.String(), "one") {
		t.Fatalf("expected body to contain only the author's chirp, got %q", w.Body.String())
	}
}

func TestHandlerGetAllChirpsInvalidAuthorID(t *testing.T) {
	store := &fakeStore{allChirps: []database.Chirp{{Body: "one"}}}
	cfg := &chirps.Config{DB: store}

	req := httptest.NewRequest("GET", "/api/chirps?author_id=not-a-uuid", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetAllChirps(w, req)

	if w.Code != 400 {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandlerGetChirpNotFound(t *testing.T) {
	store := &fakeStore{singleErr: sql.ErrNoRows}
	cfg := &chirps.Config{DB: store}

	req := httptest.NewRequest("GET", "/api/chirps/"+uuid.New().String(), nil)
	req.SetPathValue("chirpID", uuid.New().String())
	w := httptest.NewRecorder()

	cfg.HandlerGetChirp(w, req)

	if w.Code != 404 {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestHandlerGetChirpDBError(t *testing.T) {
	store := &fakeStore{singleErr: errors.New("boom")}
	cfg := &chirps.Config{DB: store}

	req := httptest.NewRequest("GET", "/api/chirps/"+uuid.New().String(), nil)
	req.SetPathValue("chirpID", uuid.New().String())
	w := httptest.NewRecorder()

	cfg.HandlerGetChirp(w, req)

	if w.Code != 500 {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestHandlerGetChirpSuccess(t *testing.T) {
	store := &fakeStore{singleChirp: database.Chirp{Body: "hello"}}
	cfg := &chirps.Config{DB: store}

	req := httptest.NewRequest("GET", "/api/chirps/"+uuid.New().String(), nil)
	req.SetPathValue("chirpID", uuid.New().String())
	w := httptest.NewRecorder()

	cfg.HandlerGetChirp(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "hello") {
		t.Fatalf("expected body to contain the chirp, got %q", w.Body.String())
	}
}

func TestHandlerDeleteChirpNoAuthHeader(t *testing.T) {
	cfg := &chirps.Config{DB: &fakeStore{}, JWTSecret: "test-secret"}

	req := httptest.NewRequest("DELETE", "/api/chirps/"+uuid.New().String(), nil)
	req.SetPathValue("chirpID", uuid.New().String())
	w := httptest.NewRecorder()

	cfg.HandlerDeleteChirp(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerDeleteChirpNotFound(t *testing.T) {
	store := &fakeStore{singleErr: sql.ErrNoRows}
	cfg := &chirps.Config{DB: store, JWTSecret: "test-secret"}
	token := signTestToken(t, uuid.New(), "test-secret", time.Hour)

	req := httptest.NewRequest("DELETE", "/api/chirps/"+uuid.New().String(), nil)
	req.SetPathValue("chirpID", uuid.New().String())
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	cfg.HandlerDeleteChirp(w, req)

	if w.Code != 404 {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
	if store.deleteCalled {
		t.Fatal("expected DeleteChirp not to be called for a missing chirp")
	}
}

func TestHandlerDeleteChirpNotOwner(t *testing.T) {
	ownerID := uuid.New()
	store := &fakeStore{singleChirp: database.Chirp{ID: uuid.New(), UserID: ownerID}}
	cfg := &chirps.Config{DB: store, JWTSecret: "test-secret"}
	token := signTestToken(t, uuid.New(), "test-secret", time.Hour)

	req := httptest.NewRequest("DELETE", "/api/chirps/"+uuid.New().String(), nil)
	req.SetPathValue("chirpID", uuid.New().String())
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	cfg.HandlerDeleteChirp(w, req)

	if w.Code != 403 {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
	if store.deleteCalled {
		t.Fatal("expected DeleteChirp not to be called when the chirp isn't owned by the caller")
	}
}

func TestHandlerDeleteChirpSuccess(t *testing.T) {
	userID := uuid.New()
	chirpID := uuid.New()
	store := &fakeStore{singleChirp: database.Chirp{ID: chirpID, UserID: userID}}
	cfg := &chirps.Config{DB: store, JWTSecret: "test-secret"}
	token := signTestToken(t, userID, "test-secret", time.Hour)

	req := httptest.NewRequest("DELETE", "/api/chirps/"+chirpID.String(), nil)
	req.SetPathValue("chirpID", chirpID.String())
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	cfg.HandlerDeleteChirp(w, req)

	if w.Code != 204 {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
	if !store.deleteCalled {
		t.Fatal("expected DeleteChirp to be called")
	}
	if store.deletedID != chirpID {
		t.Fatalf("expected chirp %v to be deleted, got %v", chirpID, store.deletedID)
	}
}
