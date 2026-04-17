package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/andreypisarev/secret-santa/internal/auth"
	"github.com/andreypisarev/secret-santa/internal/config"
	"github.com/andreypisarev/secret-santa/internal/db/sqlc"
	"github.com/andreypisarev/secret-santa/internal/email"
	mw "github.com/andreypisarev/secret-santa/internal/http/middleware"
)

type AuthHandler struct {
	Queries *sqlc.Queries
	Email   email.Sender
	Config  *config.Config
}

func (h *AuthHandler) RequestLink(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "неверный формат запроса")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		writeError(w, http.StatusBadRequest, "invalid_input", "некорректный email")
		return
	}

	token, err := auth.GenerateToken()
	if err != nil {
		slog.Error("generate token", "error", err)
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	expiresAt := time.Now().Add(15 * time.Minute)
	if err := h.Queries.CreateMagicLink(r.Context(), sqlc.CreateMagicLinkParams{
		Token:     token,
		Email:     req.Email,
		ExpiresAt: expiresAt,
	}); err != nil {
		slog.Error("create magic link", "error", err)
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	link := fmt.Sprintf("%s/api/auth/verify?token=%s", h.Config.BaseURL, token)
	html := fmt.Sprintf(`<p>Привет! Вот твоя ссылка для входа в Тайный Санта:</p><p><a href="%s">Войти</a></p><p>Ссылка действительна 15 минут.</p>`, link)

	if err := h.Email.Send(req.Email, "Вход в Тайный Санта", html); err != nil {
		slog.Error("send email", "error", err)
	}

	// Всегда 204 — не раскрываем наличие email
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "токен отсутствует")
		return
	}

	ml, err := h.Queries.GetMagicLink(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "ссылка недействительна или истекла")
		return
	}

	if err := h.Queries.MarkMagicLinkUsed(r.Context(), token); err != nil {
		slog.Error("mark magic link used", "error", err)
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	user, err := h.Queries.GetUserByEmail(r.Context(), ml.Email)
	if errors.Is(err, sql.ErrNoRows) {
		user, err = h.Queries.CreateUser(r.Context(), sqlc.CreateUserParams{
			Email: ml.Email,
			Name:  "",
		})
	}
	if err != nil {
		slog.Error("get or create user", "error", err)
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	sessionToken, err := auth.GenerateToken()
	if err != nil {
		slog.Error("generate session token", "error", err)
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	if err := h.Queries.CreateSession(r.Context(), sqlc.CreateSessionParams{
		Token:     sessionToken,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	}); err != nil {
		slog.Error("create session", "error", err)
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "s",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   !h.Config.IsDev(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 60 * 60,
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("s")
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_ = h.Queries.DeleteSession(r.Context(), cookie.Value)

	http.SetCookie(w, &http.Cookie{
		Name:     "s",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "требуется авторизация")
		return
	}

	user, err := h.Queries.GetUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"name":    user.Name,
	})
}

// WithSession оборачивает handler в RequireSession middleware (удобно для тестов).
func WithSession(queries *sqlc.Queries, next http.Handler) http.Handler {
	return mw.RequireSession(queries)(next)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   code,
		"message": message,
	})
}
