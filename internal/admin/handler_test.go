package admin_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jacobrluttrull/chirpy/internal/admin"
)

type fakeUserDeleter struct {
	err    error
	called bool
}

func (f *fakeUserDeleter) DeleteAllUsers(ctx context.Context) error {
	f.called = true
	return f.err
}

func TestFileserverHitsHandlerReportsCount(t *testing.T) {
	cfg := &admin.Config{}
	mw := cfg.MiddlewareMetricsInc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	for i := 0; i < 3; i++ {
		mw.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/app/", nil))
	}

	w := httptest.NewRecorder()
	cfg.FileserverHitsHandler(w, httptest.NewRequest("GET", "/admin/metrics", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "visited 3 times") {
		t.Fatalf("expected body to report 3 hits, got %q", w.Body.String())
	}
}

func TestResetFileserverHitsHandlerForbiddenOutsideDev(t *testing.T) {
	deleter := &fakeUserDeleter{}
	cfg := &admin.Config{DB: deleter, Platform: "production"}

	w := httptest.NewRecorder()
	cfg.ResetFileserverHitsHandler(w, httptest.NewRequest("POST", "/admin/reset", nil))

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
	if deleter.called {
		t.Fatal("expected DeleteAllUsers not to be called outside dev platform")
	}
}

func TestResetFileserverHitsHandlerResetsInDev(t *testing.T) {
	deleter := &fakeUserDeleter{}
	cfg := &admin.Config{DB: deleter, Platform: "dev"}
	cfg.MiddlewareMetricsInc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/app/", nil))

	w := httptest.NewRecorder()
	cfg.ResetFileserverHitsHandler(w, httptest.NewRequest("POST", "/admin/reset", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !deleter.called {
		t.Fatal("expected DeleteAllUsers to be called in dev platform")
	}

	hitsW := httptest.NewRecorder()
	cfg.FileserverHitsHandler(hitsW, httptest.NewRequest("GET", "/admin/metrics", nil))
	if !strings.Contains(hitsW.Body.String(), "visited 0 times") {
		t.Fatalf("expected hits counter to be reset to 0, got %q", hitsW.Body.String())
	}
}

func TestResetFileserverHitsHandlerDBError(t *testing.T) {
	deleter := &fakeUserDeleter{err: errors.New("boom")}
	cfg := &admin.Config{DB: deleter, Platform: "dev"}

	w := httptest.NewRecorder()
	cfg.ResetFileserverHitsHandler(w, httptest.NewRequest("POST", "/admin/reset", nil))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
