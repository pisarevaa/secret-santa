package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andreypisarev/secret-santa/internal/db"
	"github.com/andreypisarev/secret-santa/internal/db/sqlc"
	mw "github.com/andreypisarev/secret-santa/internal/http/middleware"
)

func setupTestDB(t *testing.T) *sqlc.Queries {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return sqlc.New(database)
}

func TestRequireSession_NoCookie(t *testing.T) {
	queries := setupTestDB(t)
	handler := mw.RequireSession(queries)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRequireSession_InvalidToken(t *testing.T) {
	queries := setupTestDB(t)
	handler := mw.RequireSession(queries)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "s", Value: "invalid-token"})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
