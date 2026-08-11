package login

import (
	"testing"
	"time"
)

func TestResolveExpirationDefaultsWhenUnspecified(t *testing.T) {
	got := resolveExpiration(0)
	if got != maxExpiration {
		t.Fatalf("expected default of %v, got %v", maxExpiration, got)
	}
}

func TestResolveExpirationDefaultsWhenNegative(t *testing.T) {
	got := resolveExpiration(-5)
	if got != maxExpiration {
		t.Fatalf("expected default of %v, got %v", maxExpiration, got)
	}
}

func TestResolveExpirationUsesRequestedWhenBelowMax(t *testing.T) {
	got := resolveExpiration(2)
	want := 2 * time.Second
	if got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestResolveExpirationCapsAtMax(t *testing.T) {
	got := resolveExpiration(7200) // 2 hours
	if got != maxExpiration {
		t.Fatalf("expected capped value of %v, got %v", maxExpiration, got)
	}
}
