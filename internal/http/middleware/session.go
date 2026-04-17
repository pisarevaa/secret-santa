package middleware

import (
	"context"
	"net/http"

	"github.com/andreypisarev/secret-santa/internal/db/sqlc"
)

type contextKey string

const UserIDKey contextKey = "userID"

func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(UserIDKey).(int64)
	return id, ok
}

// RequireSession проверяет сессионную cookie и кладет userID в контекст.
func RequireSession(queries *sqlc.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("s")
			if err != nil {
				http.Error(w, `{"error":"unauthorized","message":"требуется авторизация"}`, http.StatusUnauthorized)
				return
			}

			session, err := queries.GetSession(r.Context(), cookie.Value)
			if err != nil {
				http.Error(w, `{"error":"unauthorized","message":"сессия недействительна"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, session.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalSession — как RequireSession, но не блокирует запрос без сессии.
func OptionalSession(queries *sqlc.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("s")
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			session, err := queries.GetSession(r.Context(), cookie.Value)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, session.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
