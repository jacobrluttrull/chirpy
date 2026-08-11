package chirps

import "testing"

func TestCensorProfaneWordsReplacesKnownWords(t *testing.T) {
	got := censorProfaneWords("This is a kerfuffle opinion I need to share")
	want := "This is a **** opinion I need to share"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestCensorProfaneWordsIsCaseInsensitive(t *testing.T) {
	got := censorProfaneWords("SHARBERT and Fornax are bad words")
	want := "**** and **** are bad words"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestCensorProfaneWordsLeavesCleanBodyUnchanged(t *testing.T) {
	body := "This is a perfectly clean chirp"
	if got := censorProfaneWords(body); got != body {
		t.Fatalf("expected body to be unchanged, got %q", got)
	}
}
