package main

import (
	"bytes"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

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
	cfg := &apiConfig{jwtSecret: "test-secret"}

	req := httptest.NewRequest("POST", "/api/chirps", bytes.NewBufferString(`{"body":"hello"}`))
	w := httptest.NewRecorder()

	cfg.handlerCreateChirp(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerCreateChirpMalformedAuthHeader(t *testing.T) {
	cfg := &apiConfig{jwtSecret: "test-secret"}

	req := httptest.NewRequest("POST", "/api/chirps", bytes.NewBufferString(`{"body":"hello"}`))
	req.Header.Set("Authorization", "not-a-bearer-header")
	w := httptest.NewRecorder()

	cfg.handlerCreateChirp(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerCreateChirpInvalidToken(t *testing.T) {
	cfg := &apiConfig{jwtSecret: "test-secret"}

	req := httptest.NewRequest("POST", "/api/chirps", bytes.NewBufferString(`{"body":"hello"}`))
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	w := httptest.NewRecorder()

	cfg.handlerCreateChirp(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerCreateChirpWrongSecret(t *testing.T) {
	cfg := &apiConfig{jwtSecret: "right-secret"}
	token := signTestToken(t, uuid.New(), "wrong-secret", time.Hour)

	req := httptest.NewRequest("POST", "/api/chirps", bytes.NewBufferString(`{"body":"hello"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	cfg.handlerCreateChirp(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerCreateChirpExpiredToken(t *testing.T) {
	cfg := &apiConfig{jwtSecret: "test-secret"}
	token := signTestToken(t, uuid.New(), "test-secret", -time.Hour)

	req := httptest.NewRequest("POST", "/api/chirps", bytes.NewBufferString(`{"body":"hello"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	cfg.handlerCreateChirp(w, req)

	if w.Code != 401 {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}
