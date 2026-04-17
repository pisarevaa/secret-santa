package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andreypisarev/secret-santa/internal/config"
	"github.com/andreypisarev/secret-santa/internal/db"
	"github.com/andreypisarev/secret-santa/internal/db/sqlc"
	"github.com/andreypisarev/secret-santa/internal/email"
	"github.com/andreypisarev/secret-santa/internal/http/handlers"
)

func setupAuth(t *testing.T) *handlers.AuthHandler {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return &handlers.AuthHandler{
		Queries: sqlc.New(database),
		Email:   &email.LogSender{},
		Config: &config.Config{
			BaseURL: "http://localhost:5173",
			Env:     "development",
		},
	}
}

func TestRequestLink(t *testing.T) {
	h := setupAuth(t)

	body := strings.NewReader(`{"email":"test@example.com"}`)
	req := httptest.NewRequest("POST", "/api/auth/request-link", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.RequestLink(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestRequestLink_InvalidEmail(t *testing.T) {
	h := setupAuth(t)

	body := strings.NewReader(`{"email":"invalid"}`)
	req := httptest.NewRequest("POST", "/api/auth/request-link", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.RequestLink(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestVerify_InvalidToken(t *testing.T) {
	h := setupAuth(t)

	req := httptest.NewRequest("GET", "/api/auth/verify?token=invalid", nil)
	w := httptest.NewRecorder()

	h.Verify(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestMe_Unauthorized(t *testing.T) {
	h := setupAuth(t)

	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	w := httptest.NewRecorder()

	h.Me(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestLogout(t *testing.T) {
	h := setupAuth(t)

	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "s", Value: "some-token"})
	w := httptest.NewRecorder()

	h.Logout(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}

	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "s" && c.MaxAge < 0 {
			found = true
		}
	}
	if !found {
		t.Error("session cookie not cleared")
	}
}

func TestLogout_NoCookie(t *testing.T) {
	h := setupAuth(t)

	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	w := httptest.NewRecorder()

	h.Logout(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

// testEmailSender — записывает отправленные письма для проверки в тестах.
type testEmailSender struct {
	Sent []struct{ To, Subject, HTML string }
}

func (s *testEmailSender) Send(to, subject, html string) error {
	s.Sent = append(s.Sent, struct{ To, Subject, HTML string }{to, subject, html})
	return nil
}

func TestVerify_FullFlow(t *testing.T) {
	h := setupAuth(t)
	sender := &testEmailSender{}
	h.Email = sender

	body := strings.NewReader(`{"email":"flow@example.com"}`)
	req := httptest.NewRequest("POST", "/api/auth/request-link", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.RequestLink(w, req)

	if len(sender.Sent) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(sender.Sent))
	}

	html := sender.Sent[0].HTML
	prefix := h.Config.BaseURL + "/api/auth/verify?token="
	idx := strings.Index(html, prefix)
	if idx == -1 {
		t.Fatalf("link not found in email: %s", html)
	}
	tokenStart := idx + len(prefix)
	tokenEnd := strings.Index(html[tokenStart:], `"`)
	token := html[tokenStart : tokenStart+tokenEnd]

	req = httptest.NewRequest("GET", "/api/auth/verify?token="+token, nil)
	w = httptest.NewRecorder()
	h.Verify(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("verify: status = %d, want %d", w.Code, http.StatusFound)
	}

	var sessionCookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "s" {
			sessionCookie = c
		}
	}
	if sessionCookie == nil {
		t.Fatal("session cookie not set")
	}

	req = httptest.NewRequest("GET", "/api/auth/me", nil)
	req.AddCookie(sessionCookie)
	w = httptest.NewRecorder()

	mwHandler := handlers.WithSession(h.Queries, http.HandlerFunc(h.Me))
	mwHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("me: status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var meResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&meResp)
	if meResp["email"] != "flow@example.com" {
		t.Errorf("email = %v, want flow@example.com", meResp["email"])
	}
}
