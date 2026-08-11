package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/jacobrluttrull/chirpy/internal/auth"
)

// signToken builds a JWT the same way the (unexported) makeJWT helper does,
// so ValidateJWT can be exercised as a black box from outside the package.
func signToken(t *testing.T, userID uuid.UUID, secret string, expiresIn time.Duration) string {
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

func TestValidateJWTValidToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	tokenString := signToken(t, userID, secret, time.Hour)

	gotID, err := auth.ValidateJWT(tokenString, secret)
	if err != nil {
		t.Fatalf("ValidateJWT returned error: %v", err)
	}
	if gotID != userID {
		t.Fatalf("expected user ID %v, got %v", userID, gotID)
	}
}

func TestValidateJWTWrongSecret(t *testing.T) {
	tokenString := signToken(t, uuid.New(), "right-secret", time.Hour)

	_, err := auth.ValidateJWT(tokenString, "wrong-secret")
	if err == nil {
		t.Fatal("expected error when validating with wrong secret")
	}
}

func TestValidateJWTExpired(t *testing.T) {
	tokenString := signToken(t, uuid.New(), "test-secret", -time.Hour)

	_, err := auth.ValidateJWT(tokenString, "test-secret")
	if err == nil {
		t.Fatal("expected error when validating expired token")
	}
}

func TestValidateJWTMalformed(t *testing.T) {
	_, err := auth.ValidateJWT("not-a-valid-token", "test-secret")
	if err == nil {
		t.Fatal("expected error when validating a malformed token")
	}
}
