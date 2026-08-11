package response_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/jacobrluttrull/chirpy/internal/response"
)

func TestJSONWritesStatusAndBody(t *testing.T) {
	w := httptest.NewRecorder()

	response.JSON(w, 201, map[string]string{"hello": "world"})

	if w.Code != 201 {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if got["hello"] != "world" {
		t.Fatalf("expected body to contain hello=world, got %v", got)
	}
}

func TestErrorWritesErrorShape(t *testing.T) {
	w := httptest.NewRecorder()

	response.Error(w, 400, "bad request")

	if w.Code != 400 {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var got struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if got.Error != "bad request" {
		t.Fatalf("expected error message %q, got %q", "bad request", got.Error)
	}
}
