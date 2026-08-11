package auth_test

import (
	"testing"

	"github.com/jacobrluttrull/chirpy/internal/auth"
)

func TestHashPasswordAndCheck(t *testing.T) {
	password := "correct-horse-battery-staple"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	if hash == password {
		t.Fatal("hash should not equal the plaintext password")
	}

	match, err := auth.CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash returned error: %v", err)
	}
	if !match {
		t.Fatal("expected correct password to match hash")
	}
}

func TestCheckPasswordHashWrongPassword(t *testing.T) {
	hash, err := auth.HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	match, err := auth.CheckPasswordHash("wrong-password", hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash returned error: %v", err)
	}
	if match {
		t.Fatal("expected wrong password not to match hash")
	}
}
