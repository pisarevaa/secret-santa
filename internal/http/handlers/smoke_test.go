package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/andreypisarev/secret-santa/internal/config"
	"github.com/andreypisarev/secret-santa/internal/db"
	"github.com/andreypisarev/secret-santa/internal/db/sqlc"
	"github.com/andreypisarev/secret-santa/internal/http/handlers"
	mw "github.com/andreypisarev/secret-santa/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

func setupServer(t *testing.T) (*chi.Mux, *testEmailSender) {
	t.Helper()

	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	queries := sqlc.New(database)
	sender := &testEmailSender{}
	cfg := &config.Config{BaseURL: "http://localhost", Env: "development"}

	authH := &handlers.AuthHandler{Queries: queries, Email: sender, Config: cfg}
	groupH := &handlers.GroupHandler{Queries: queries}
	drawH := &handlers.DrawHandler{Queries: queries, DB: database}

	r := chi.NewRouter()
	r.Post("/api/auth/request-link", authH.RequestLink)
	r.Get("/api/auth/verify", authH.Verify)

	r.Group(func(r chi.Router) {
		r.Use(mw.RequireSession(queries))
		r.Get("/api/auth/me", authH.Me)
		r.Post("/api/groups", groupH.Create)
		r.Post("/api/groups/{inviteCode}/join", groupH.Join)
		r.Post("/api/groups/{id}/draw", drawH.Draw)
		r.Get("/api/groups/{id}/my-recipient", drawH.MyRecipient)
	})

	r.Group(func(r chi.Router) {
		r.Use(mw.OptionalSession(queries))
		r.Get("/api/groups/{inviteCode}", groupH.GetByInviteCode)
	})

	return r, sender
}

// login выполняет полный flow аутентификации и возвращает session cookie.
func login(t *testing.T, r *chi.Mux, sender *testEmailSender, emailAddr string) *http.Cookie {
	t.Helper()

	body := strings.NewReader(`{"email":"` + emailAddr + `"}`)
	req := httptest.NewRequest("POST", "/api/auth/request-link", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("request-link: %d", w.Code)
	}

	html := sender.Sent[len(sender.Sent)-1].HTML
	prefix := "http://localhost/api/auth/verify?token="
	idx := strings.Index(html, prefix)
	if idx == -1 {
		t.Fatal("token not in email")
	}
	start := idx + len(prefix)
	end := strings.Index(html[start:], `"`)
	token := html[start : start+end]

	req = httptest.NewRequest("GET", "/api/auth/verify?token="+token, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("verify: %d, body: %s", w.Code, w.Body.String())
	}

	for _, c := range w.Result().Cookies() {
		if c.Name == "s" {
			return c
		}
	}
	t.Fatal("no session cookie")
	return nil
}

func TestSmoke_FullFlow(t *testing.T) {
	r, sender := setupServer(t)

	// 1. Организатор логинится
	orgCookie := login(t, r, sender, "org@test.com")

	// 2. Создает группу
	body := strings.NewReader(`{"title":"Новый год"}`)
	req := httptest.NewRequest("POST", "/api/groups", body)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(orgCookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create group: %d, body: %s", w.Code, w.Body.String())
	}

	var groupResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&groupResp)
	inviteCode := groupResp["invite_code"].(string)
	groupID := int(groupResp["id"].(float64))

	// 3. Организатор вступает
	body = strings.NewReader(`{"name":"Организатор","wishlist":"Виски"}`)
	req = httptest.NewRequest("POST", "/api/groups/"+inviteCode+"/join", body)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(orgCookie)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("org join: %d, body: %s", w.Code, w.Body.String())
	}

	// 4. Два участника логинятся и вступают
	for i, emailAddr := range []string{"alice@test.com", "bob@test.com"} {
		cookie := login(t, r, sender, emailAddr)
		names := []string{"Алиса", "Боб"}

		body = strings.NewReader(`{"name":"` + names[i] + `","wishlist":"Подарок"}`)
		req = httptest.NewRequest("POST", "/api/groups/"+inviteCode+"/join", body)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(cookie)
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("join %s: %d", emailAddr, w.Code)
		}
	}

	// 5. Организатор проводит жеребьевку
	req = httptest.NewRequest("POST", "/api/groups/"+strconv.Itoa(groupID)+"/draw", nil)
	req.AddCookie(orgCookie)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("draw: %d, body: %s", w.Code, w.Body.String())
	}

	// 6. Организатор видит своего подопечного
	req = httptest.NewRequest("GET", "/api/groups/"+strconv.Itoa(groupID)+"/my-recipient", nil)
	req.AddCookie(orgCookie)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("my-recipient: %d, body: %s", w.Code, w.Body.String())
	}

	var recipientResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&recipientResp)
	recipient := recipientResp["recipient"].(map[string]interface{})
	if recipient["name"] == "" {
		t.Error("recipient name is empty")
	}

	t.Logf("Организатор дарит: %s", recipient["name"])
}
