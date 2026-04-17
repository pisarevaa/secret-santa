package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/andreypisarev/secret-santa/internal/db"
	"github.com/andreypisarev/secret-santa/internal/db/sqlc"
	"github.com/andreypisarev/secret-santa/internal/http/handlers"
	"github.com/go-chi/chi/v5"
)

func setupDraw(t *testing.T) (*handlers.DrawHandler, *sqlc.Queries) {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	queries := sqlc.New(database)
	return &handlers.DrawHandler{Queries: queries, DB: database}, queries
}

func TestDraw(t *testing.T) {
	h, queries := setupDraw(t)

	org, _ := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		Email: "org@test.com", Name: "Org",
	})
	group, _ := queries.CreateGroup(context.Background(), sqlc.CreateGroupParams{
		InviteCode: "draw12345678", Title: "Тест", OrganizerID: org.ID,
	})

	for i := 0; i < 3; i++ {
		u, _ := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
			Email: "u" + strconv.Itoa(i) + "@test.com", Name: "User" + strconv.Itoa(i),
		})
		queries.CreateMembership(context.Background(), sqlc.CreateMembershipParams{
			GroupID: group.ID, UserID: u.ID, Wishlist: "Подарок",
		})
	}

	queries.CreateMembership(context.Background(), sqlc.CreateMembershipParams{
		GroupID: group.ID, UserID: org.ID, Wishlist: "Мой подарок",
	})

	r := chi.NewRouter()
	r.Post("/api/groups/{id}/draw", h.Draw)

	req := httptest.NewRequest("POST", "/api/groups/"+strconv.FormatInt(group.ID, 10)+"/draw", nil)
	req = withUserID(req, org.ID)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body: %s", w.Code, http.StatusNoContent, w.Body.String())
	}

	updatedGroup, _ := queries.GetGroupByID(context.Background(), group.ID)
	if updatedGroup.Status != "drawn" {
		t.Errorf("group status = %q, want %q", updatedGroup.Status, "drawn")
	}
	if !updatedGroup.DrawnAt.Valid {
		t.Error("drawn_at is NULL, want set")
	}

	members, _ := queries.ListMembershipsByGroup(context.Background(), group.ID)
	recipients := make(map[int64]bool)
	for _, m := range members {
		if !m.RecipientID.Valid {
			t.Errorf("member %d has no recipient", m.UserID)
			continue
		}
		if m.RecipientID.Int64 == m.UserID {
			t.Errorf("member %d assigned to self", m.UserID)
		}
		if recipients[m.RecipientID.Int64] {
			t.Errorf("recipient %d assigned twice", m.RecipientID.Int64)
		}
		recipients[m.RecipientID.Int64] = true
	}
}

func TestDraw_NotOrganizer(t *testing.T) {
	h, queries := setupDraw(t)

	org, _ := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		Email: "org@test.com", Name: "Org",
	})
	other, _ := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		Email: "other@test.com", Name: "Other",
	})
	group, _ := queries.CreateGroup(context.Background(), sqlc.CreateGroupParams{
		InviteCode: "notorg123456", Title: "Тест", OrganizerID: org.ID,
	})

	r := chi.NewRouter()
	r.Post("/api/groups/{id}/draw", h.Draw)

	req := httptest.NewRequest("POST", "/api/groups/"+strconv.FormatInt(group.ID, 10)+"/draw", nil)
	req = withUserID(req, other.ID)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestDraw_AlreadyDrawn(t *testing.T) {
	h, queries := setupDraw(t)

	org, _ := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		Email: "org@test.com", Name: "Org",
	})
	group, _ := queries.CreateGroup(context.Background(), sqlc.CreateGroupParams{
		InviteCode: "already12345", Title: "Тест", OrganizerID: org.ID,
	})
	queries.CreateMembership(context.Background(), sqlc.CreateMembershipParams{
		GroupID: group.ID, UserID: org.ID, Wishlist: "a",
	})
	u, _ := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		Email: "p@test.com", Name: "P",
	})
	queries.CreateMembership(context.Background(), sqlc.CreateMembershipParams{
		GroupID: group.ID, UserID: u.ID, Wishlist: "b",
	})

	r := chi.NewRouter()
	r.Post("/api/groups/{id}/draw", h.Draw)

	req1 := httptest.NewRequest("POST", "/api/groups/"+strconv.FormatInt(group.ID, 10)+"/draw", nil)
	req1 = withUserID(req1, org.ID)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusNoContent {
		t.Fatalf("first draw: status = %d, want %d", w1.Code, http.StatusNoContent)
	}

	req2 := httptest.NewRequest("POST", "/api/groups/"+strconv.FormatInt(group.ID, 10)+"/draw", nil)
	req2 = withUserID(req2, org.ID)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusConflict {
		t.Errorf("second draw: status = %d, want %d", w2.Code, http.StatusConflict)
	}
}

func TestDraw_NotEnoughMembers(t *testing.T) {
	h, queries := setupDraw(t)

	org, _ := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		Email: "org@test.com", Name: "Org",
	})
	group, _ := queries.CreateGroup(context.Background(), sqlc.CreateGroupParams{
		InviteCode: "tooFew123456", Title: "Тест", OrganizerID: org.ID,
	})
	queries.CreateMembership(context.Background(), sqlc.CreateMembershipParams{
		GroupID: group.ID, UserID: org.ID, Wishlist: "a",
	})

	r := chi.NewRouter()
	r.Post("/api/groups/{id}/draw", h.Draw)

	req := httptest.NewRequest("POST", "/api/groups/"+strconv.FormatInt(group.ID, 10)+"/draw", nil)
	req = withUserID(req, org.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestMyRecipient(t *testing.T) {
	h, queries := setupDraw(t)

	org, _ := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		Email: "org@test.com", Name: "Org",
	})
	group, _ := queries.CreateGroup(context.Background(), sqlc.CreateGroupParams{
		InviteCode: "myrecipient1", Title: "Тест", OrganizerID: org.ID,
	})
	queries.CreateMembership(context.Background(), sqlc.CreateMembershipParams{
		GroupID: group.ID, UserID: org.ID, Wishlist: "org wish",
	})
	u, _ := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		Email: "p@test.com", Name: "Pavel",
	})
	queries.CreateMembership(context.Background(), sqlc.CreateMembershipParams{
		GroupID: group.ID, UserID: u.ID, Wishlist: "книга",
	})

	r := chi.NewRouter()
	r.Post("/api/groups/{id}/draw", h.Draw)
	r.Get("/api/groups/{id}/my-recipient", h.MyRecipient)

	drawReq := httptest.NewRequest("POST", "/api/groups/"+strconv.FormatInt(group.ID, 10)+"/draw", nil)
	drawReq = withUserID(drawReq, org.ID)
	drawW := httptest.NewRecorder()
	r.ServeHTTP(drawW, drawReq)
	if drawW.Code != http.StatusNoContent {
		t.Fatalf("draw: status = %d, want %d", drawW.Code, http.StatusNoContent)
	}

	req := httptest.NewRequest("GET", "/api/groups/"+strconv.FormatInt(group.ID, 10)+"/my-recipient", nil)
	req = withUserID(req, org.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp struct {
		Recipient struct {
			Name     string `json:"name"`
			Wishlist string `json:"wishlist"`
		} `json:"recipient"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Recipient.Name == "" {
		t.Error("recipient name is empty")
	}
	if resp.Recipient.Wishlist == "" {
		t.Error("recipient wishlist is empty")
	}
}
