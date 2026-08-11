package auth_test

import (
	"net/http"
	"testing"

	"github.com/jacobrluttrull/chirpy/internal/auth"
)

func TestGetBearerTokenValid(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer some-token-value")

	token, err := auth.GetBearerToken(headers)
	if err != nil {
		t.Fatalf("GetBearerToken returned error: %v", err)
	}
	if token != "some-token-value" {
		t.Fatalf("expected token %q, got %q", "some-token-value", token)
	}
}

func TestGetBearerTokenMissingHeader(t *testing.T) {
	headers := http.Header{}

	_, err := auth.GetBearerToken(headers)
	if err == nil {
		t.Fatal("expected error when Authorization header is missing")
	}
}

func TestGetBearerTokenMissingScheme(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "some-token-value")

	_, err := auth.GetBearerToken(headers)
	if err == nil {
		t.Fatal("expected error when Authorization header has no scheme")
	}
}

func TestGetBearerTokenWrongScheme(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Basic some-token-value")

	_, err := auth.GetBearerToken(headers)
	if err == nil {
		t.Fatal("expected error when Authorization header uses the wrong scheme")
	}
}
