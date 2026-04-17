package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andreypisarev/secret-santa/internal/db"
	"github.com/andreypisarev/secret-santa/internal/db/sqlc"
	"github.com/andreypisarev/secret-santa/internal/http/handlers"
	mw "github.com/andreypisarev/secret-santa/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

func setupGroups(t *testing.T) (*handlers.GroupHandler, *sqlc.Queries) {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	queries := sqlc.New(database)
	return &handlers.GroupHandler{Queries: queries}, queries
}

func withUserID(r *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(r.Context(), mw.UserIDKey, userID)
	return r.WithContext(ctx)
}

func TestCreateGroup(t *testing.T) {
	h, queries := setupGroups(t)

	user, _ := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		Email: "org@example.com",
		Name:  "Организатор",
	})

	body := strings.NewReader(`{"title":"Новый год 2026"}`)
	req := httptest.NewRequest("POST", "/api/groups", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, user.ID)
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
	}
}

func TestCreateGroup_EmptyTitle(t *testing.T) {
	h, queries := setupGroups(t)
	user, _ := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		Email: "org@example.com",
		Name:  "Организатор",
	})

	body := strings.NewReader(`{"title":""}`)
	req := httptest.NewRequest("POST", "/api/groups", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, user.ID)
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestJoinGroup(t *testing.T) {
	h, queries := setupGroups(t)

	org, _ := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		Email: "org@example.com", Name: "Организатор",
	})
	group, _ := queries.CreateGroup(context.Background(), sqlc.CreateGroupParams{
		InviteCode: "testcode1234", Title: "Тест", OrganizerID: org.ID,
	})

	member, _ := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		Email: "member@example.com", Name: "",
	})

	r := chi.NewRouter()
	r.Post("/api/groups/{inviteCode}/join", h.Join)

	body := strings.NewReader(`{"name":"Вася","wishlist":"Книга"}`)
	req := httptest.NewRequest("POST", "/api/groups/"+group.InviteCode+"/join", body)
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, member.ID)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d, body: %s", w.Code, http.StatusNoContent, w.Body.String())
	}
}
