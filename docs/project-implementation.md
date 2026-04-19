# Тайный Санта — План реализации

> **Для агентов:** ОБЯЗАТЕЛЬНЫЙ НАВЫК: Используй superpowers:subagent-driven-development (рекомендуется) или superpowers:executing-plans для выполнения плана задача за задачей. Шаги используют синтаксис `- [ ]` для отслеживания.

**Цель:** Веб-приложение для игры «Тайный Санта» — организатор создает группу, участники регистрируются через magic link, после жеребьевки каждый видит подопечного и может анонимно переписываться.

**Архитектура:** Go-монолит обслуживает REST API, WebSocket и встроенный React-фронт с одного домена. SQLite в файле, один инстанс на Fly.io. Пакеты в `internal/` содержат чистую доменную логику без зависимости от HTTP.

**Стек:** Go 1.22+ (chi, coder/websocket, modernc.org/sqlite, sqlc, golang-migrate) | React 18 + Vite + TypeScript + Tailwind CSS v4 + react-router v7

---

## Процесс выполнения

**Изучи контекст:**
- Прочитай `docs/project-design.md` — полный дизайн проекта
- Найди первую незакрытую фичу в этом плане

**Выполни фичу:**
- Выполни все шаги фичи
- Тесты обязательны для каждой фичи

**Подведи итог:**
- Напиши короткий итог: что сделал, что работает, есть ли проблемы
- Жди подтверждения пользователя
- После подтверждения: отметь шаги `- [x]` и поставь ✅ в заголовке фичи

## Принципы

- **DRY** — не дублируй код, выноси общее
- **YAGNI** — не добавляй то, что не описано в дизайне
- **TDD** — сначала тест, потом реализация
- **UI на русском** — весь интерфейс на русском языке, без буквы «е»... то есть без «ё»

---

### Фича 1: Инициализация проекта и структура ✅

**Цель:** Go-модуль компилируется, Vite-проект собирается, dev-режим работает (Vite проксирует на Go), healthcheck отвечает 200.

**Файлы:**
- Создать: `go.mod`, `go.sum`
- Создать: `cmd/server/main.go` — точка входа, запуск HTTP-сервера
- Создать: `internal/config/config.go` — чтение env-переменных
- Создать: `internal/config/config_test.go`
- Создать: `embed.go` — `//go:embed web/dist`
- Создать: `web/package.json`, `web/tsconfig.json`, `web/vite.config.ts`, `web/index.html`
- Создать: `web/src/main.tsx`, `web/src/App.tsx`, `web/src/vite-env.d.ts`
- Создать: `web/src/index.css` — Tailwind
- Создать: `.env.example`
- Создать: `.gitignore`

**Шаги:**

- [x] **Шаг 1: Создать `.gitignore`**

```gitignore
# Go
/cmd/server/server

# Node
web/node_modules/
web/dist/

# Environment
.env

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store
```

- [x] **Шаг 2: Инициализировать Go-модуль**

```bash
cd /Users/andreypisarev/other/secret-santa
go mod init github.com/andreypisarev/secret-santa
```

- [x] **Шаг 3: Создать `internal/config/config.go`**

```go
package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	BaseURL      string
	DatabasePath string
	ResendAPIKey string
	EmailFrom    string
	Port         int
	Env          string // "production" | "development"
	LogLevel     string
}

func Load() (*Config, error) {
	port := 8080
	if v := os.Getenv("PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT: %w", err)
		}
		port = p
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "app.db"
	}

	return &Config{
		BaseURL:      os.Getenv("BASE_URL"),
		DatabasePath: dbPath,
		ResendAPIKey: os.Getenv("RESEND_API_KEY"),
		EmailFrom:    os.Getenv("EMAIL_FROM"),
		Port:         port,
		Env:          env,
		LogLevel:     logLevel,
	}, nil
}

func (c *Config) IsDev() bool {
	return c.Env == "development"
}
```

- [x] **Шаг 4: Написать тест для config**

Создать `internal/config/config_test.go`:

```go
package config_test

import (
	"os"
	"testing"

	"github.com/andreypisarev/secret-santa/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	// Очистить переменные, которые могут быть установлены
	os.Unsetenv("PORT")
	os.Unsetenv("ENV")
	os.Unsetenv("DATABASE_PATH")
	os.Unsetenv("LOG_LEVEL")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Port)
	}
	if cfg.Env != "development" {
		t.Errorf("Env = %q, want %q", cfg.Env, "development")
	}
	if !cfg.IsDev() {
		t.Error("IsDev() = false, want true")
	}
	if cfg.DatabasePath != "app.db" {
		t.Errorf("DatabasePath = %q, want %q", cfg.DatabasePath, "app.db")
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("PORT", "3000")
	t.Setenv("ENV", "production")
	t.Setenv("DATABASE_PATH", "/data/app.db")
	t.Setenv("BASE_URL", "https://example.com")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 3000 {
		t.Errorf("Port = %d, want 3000", cfg.Port)
	}
	if cfg.Env != "production" {
		t.Errorf("Env = %q, want %q", cfg.Env, "production")
	}
	if cfg.IsDev() {
		t.Error("IsDev() = true, want false")
	}
	if cfg.DatabasePath != "/data/app.db" {
		t.Errorf("DatabasePath = %q, want %q", cfg.DatabasePath, "/data/app.db")
	}
	if cfg.BaseURL != "https://example.com" {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, "https://example.com")
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv("PORT", "abc")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for invalid PORT")
	}
}
```

- [x] **Шаг 5: Запустить тест config и убедиться, что проходит**

```bash
go test ./internal/config/ -v
```

Ожидаемый результат: все 3 теста PASS.

- [x] **Шаг 6: Создать `cmd/server/main.go`**

```go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andreypisarev/secret-santa/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: r,
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("server starting", "port", cfg.Port, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
```

- [x] **Шаг 7: Установить зависимость chi и убедиться, что сервер компилируется**

```bash
go get github.com/go-chi/chi/v5
go build ./cmd/server/
```

Ожидаемый результат: компиляция без ошибок.

- [x] **Шаг 8: Инициализировать Vite-проект с React + TypeScript**

```bash
cd /Users/andreypisarev/other/secret-santa
npm create vite@latest web -- --template react-ts
cd web
npm install
npm install -D @tailwindcss/vite
```

- [x] **Шаг 9: Настроить Tailwind CSS в Vite**

Заменить содержимое `web/vite.config.ts`:

```ts
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5173,
    proxy: {
      "/api": "http://localhost:8080",
      "/ws": {
        target: "ws://localhost:8080",
        ws: true,
      },
    },
  },
});
```

Заменить содержимое `web/src/index.css`:

```css
@import "tailwindcss";
```

- [x] **Шаг 10: Создать минимальный `web/src/App.tsx`**

```tsx
function App() {
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <h1 className="text-3xl font-bold text-gray-900">Тайный Санта</h1>
    </div>
  );
}

export default App;
```

Заменить `web/src/main.tsx`:

```tsx
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import "./index.css";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
```

- [x] **Шаг 11: Создать `embed.go` в корне проекта**

```go
package secretsanta

import "embed"

//go:embed web/dist
var WebDist embed.FS
```

- [x] **Шаг 12: Создать `.env.example`**

```
BASE_URL=http://localhost:5173
DATABASE_PATH=app.db
SESSION_SECRET=dev-secret-change-me
RESEND_API_KEY=
EMAIL_FROM=noreply@localhost
PORT=8080
ENV=development
LOG_LEVEL=debug
```

- [x] **Шаг 13: Убедиться, что фронт собирается**

```bash
cd /Users/andreypisarev/other/secret-santa/web
npm run build
```

Ожидаемый результат: `web/dist/` создана, сборка без ошибок.

- [x] **Шаг 14: Убедиться, что Go-проект компилируется с embed**

```bash
cd /Users/andreypisarev/other/secret-santa
go build ./cmd/server/
```

Ожидаемый результат: компиляция без ошибок.

- [x] **Шаг 15: Коммит**

```bash
git add .gitignore go.mod go.sum cmd/ internal/config/ embed.go web/ .env.example
git commit -m "feat: инициализация проекта — Go-сервер с chi, Vite + React + Tailwind"
```

**Проверка:**
- [x] `go test ./internal/config/ -v` — все тесты проходят
- [x] `go build ./cmd/server/` — компилируется без ошибок
- [x] `cd web && npm run build` — фронт собирается

---

### Фича 2: База данных и миграции ✅

**Цель:** SQLite подключается на старте, миграции применяются автоматически, таблицы создаются. sqlc генерирует типобезопасные Go-запросы.

**Файлы:**
- Создать: `internal/db/db.go` — подключение к SQLite
- Создать: `internal/db/db_test.go`
- Создать: `internal/db/migrations/001_init.up.sql`
- Создать: `internal/db/migrations/001_init.down.sql`
- Создать: `internal/db/queries/users.sql`
- Создать: `internal/db/queries/groups.sql`
- Создать: `internal/db/queries/memberships.sql`
- Создать: `internal/db/queries/auth.sql`
- Создать: `internal/db/queries/messages.sql`
- Создать: `sqlc.yaml`
- Создать: `internal/db/sqlc/` — генерируемый код (не редактировать вручную)
- Изменить: `cmd/server/main.go` — подключить БД

**Шаги:**

- [x] **Шаг 1: Создать миграцию `internal/db/migrations/001_init.up.sql`**

```sql
CREATE TABLE users (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    email       TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE groups (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    invite_code  TEXT NOT NULL UNIQUE,
    title        TEXT NOT NULL,
    organizer_id INTEGER NOT NULL REFERENCES users(id),
    status       TEXT NOT NULL DEFAULT 'open',
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    drawn_at     DATETIME
);

CREATE TABLE memberships (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id      INTEGER NOT NULL REFERENCES groups(id),
    user_id       INTEGER NOT NULL REFERENCES users(id),
    wishlist      TEXT NOT NULL DEFAULT '',
    recipient_id  INTEGER REFERENCES users(id),
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, user_id)
);

CREATE TABLE magic_links (
    token       TEXT PRIMARY KEY,
    email       TEXT NOT NULL,
    expires_at  DATETIME NOT NULL,
    used_at     DATETIME
);

CREATE TABLE sessions (
    token       TEXT PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id),
    expires_at  DATETIME NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE messages (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id       INTEGER NOT NULL REFERENCES groups(id),
    sender_id      INTEGER NOT NULL REFERENCES users(id),
    recipient_id   INTEGER NOT NULL REFERENCES users(id),
    direction      TEXT NOT NULL,
    body           TEXT NOT NULL,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_pair ON messages(group_id, sender_id, recipient_id, created_at);
CREATE INDEX idx_memberships_group ON memberships(group_id);
```

- [x] **Шаг 2: Создать миграцию `internal/db/migrations/001_init.down.sql`**

```sql
DROP INDEX IF EXISTS idx_memberships_group;
DROP INDEX IF EXISTS idx_messages_pair;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS magic_links;
DROP TABLE IF EXISTS memberships;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS users;
```

- [x] **Шаг 3: Создать `internal/db/db.go`**

```go
package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Open(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(wal)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// SQLite работает лучше с одним соединением для записи
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("create source driver: %w", err)
	}

	dbDriver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("create db driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite", dbDriver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}

	slog.Info("migrations applied")
	return nil
}
```

**Важно:** `golang-migrate` с SQLite через `modernc.org/sqlite` может потребовать драйвер `sqlite` вместо `sqlite3`. Если `github.com/golang-migrate/migrate/v4/database/sqlite` не существует, используй `github.com/golang-migrate/migrate/v4/database/sqlite3` — он работает с любым драйвером, зарегистрированным как `"sqlite"` или `"sqlite3"` в `database/sql`. Проверь при компиляции и поправь импорт.

- [x] **Шаг 4: Написать тест для подключения и миграций**

Создать `internal/db/db_test.go`:

```go
package db_test

import (
	"testing"

	"github.com/andreypisarev/secret-santa/internal/db"
)

func TestOpenAndMigrate(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	// Проверяем, что таблицы созданы
	tables := []string{"users", "groups", "memberships", "magic_links", "sessions", "messages"}
	for _, table := range tables {
		var name string
		err := database.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	// Повторный вызов не должен падать
	if err := db.Migrate(database); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
}
```

- [x] **Шаг 5: Установить зависимости и запустить тесты**

```bash
go get modernc.org/sqlite
go get github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/sqlite
go get github.com/golang-migrate/migrate/v4/source/iofs
go test ./internal/db/ -v
```

Ожидаемый результат: оба теста PASS. Если импорт `database/sqlite` не работает — заменить на `database/sqlite3` и повторить.

- [x] **Шаг 6: Создать `sqlc.yaml` в корне проекта**

```yaml
version: "2"
sql:
  - engine: "sqlite"
    queries: "internal/db/queries"
    schema: "internal/db/migrations/001_init.up.sql"
    gen:
      go:
        package: "sqlc"
        out: "internal/db/sqlc"
```

- [x] **Шаг 7: Создать SQL-запросы для sqlc — `internal/db/queries/users.sql`**

```sql
-- name: CreateUser :one
INSERT INTO users (email, name) VALUES (?, ?) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: UpdateUserName :exec
UPDATE users SET name = ? WHERE id = ?;
```

- [x] **Шаг 8: Создать `internal/db/queries/auth.sql`**

```sql
-- name: CreateMagicLink :exec
INSERT INTO magic_links (token, email, expires_at) VALUES (?, ?, ?);

-- name: GetMagicLink :one
SELECT * FROM magic_links WHERE token = ? AND used_at IS NULL AND expires_at > CURRENT_TIMESTAMP;

-- name: MarkMagicLinkUsed :exec
UPDATE magic_links SET used_at = CURRENT_TIMESTAMP WHERE token = ?;

-- name: CreateSession :exec
INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?);

-- name: GetSession :one
SELECT * FROM sessions WHERE token = ? AND expires_at > CURRENT_TIMESTAMP;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token = ?;
```

- [x] **Шаг 9: Создать `internal/db/queries/groups.sql`**

```sql
-- name: CreateGroup :one
INSERT INTO groups (invite_code, title, organizer_id, status) VALUES (?, ?, ?, 'open') RETURNING *;

-- name: GetGroupByInviteCode :one
SELECT * FROM groups WHERE invite_code = ?;

-- name: GetGroupByID :one
SELECT * FROM groups WHERE id = ?;

-- name: DrawGroup :execresult
UPDATE groups SET status = 'drawn', drawn_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'open';
```

- [x] **Шаг 10: Создать `internal/db/queries/memberships.sql`**

```sql
-- name: CreateMembership :one
INSERT INTO memberships (group_id, user_id, wishlist) VALUES (?, ?, ?) RETURNING *;

-- name: GetMembership :one
SELECT * FROM memberships WHERE id = ?;

-- name: GetMembershipByGroupAndUser :one
SELECT * FROM memberships WHERE group_id = ? AND user_id = ?;

-- name: ListMembershipsByGroup :many
SELECT * FROM memberships WHERE group_id = ?;

-- name: UpdateWishlist :exec
UPDATE memberships SET wishlist = ? WHERE id = ?;

-- name: SetRecipient :exec
UPDATE memberships SET recipient_id = ? WHERE group_id = ? AND user_id = ?;

-- name: GetMyRecipient :one
SELECT u.name, m.wishlist
FROM memberships m
JOIN users u ON u.id = m.recipient_id
WHERE m.group_id = ? AND m.user_id = ? AND m.recipient_id IS NOT NULL;

-- name: CountMembersByGroup :one
SELECT COUNT(*) FROM memberships WHERE group_id = ?;
```

- [x] **Шаг 11: Создать `internal/db/queries/messages.sql`**

```sql
-- name: CreateMessage :one
INSERT INTO messages (group_id, sender_id, recipient_id, direction, body) VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: ListMessages :many
SELECT * FROM messages
WHERE group_id = ? AND sender_id = ? AND recipient_id = ? AND direction = ?
ORDER BY created_at DESC
LIMIT 50;

-- name: ListMessagesBefore :many
SELECT * FROM messages
WHERE group_id = ? AND sender_id = ? AND recipient_id = ? AND direction = ? AND id < ?
ORDER BY created_at DESC
LIMIT 50;
```

- [x] **Шаг 12: Установить sqlc и сгенерировать код**

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
sqlc generate
```

Ожидаемый результат: файлы появятся в `internal/db/sqlc/`.

- [x] **Шаг 13: Подключить БД в `cmd/server/main.go`**

Добавить после загрузки конфига:

```go
database, err := db.Open(cfg.DatabasePath)
if err != nil {
    slog.Error("failed to open database", "error", err)
    os.Exit(1)
}
defer database.Close()

if err := db.Migrate(database); err != nil {
    slog.Error("failed to run migrations", "error", err)
    os.Exit(1)
}
```

Добавить импорт `"github.com/andreypisarev/secret-santa/internal/db"`.

- [x] **Шаг 14: Проверить компиляцию и тесты**

```bash
go build ./cmd/server/
go test ./internal/... -v
```

Ожидаемый результат: компиляция и все тесты проходят.

- [x] **Шаг 15: Коммит**

```bash
git add sqlc.yaml internal/db/ cmd/server/main.go go.mod go.sum
git commit -m "feat: SQLite + миграции + sqlc-запросы для всех таблиц"
```

**Проверка:**
- [x] `go test ./internal/db/ -v` — тесты миграций проходят
- [x] `sqlc generate` — генерация без ошибок
- [x] `go build ./cmd/server/` — компилируется

---

### Фича 3: Аутентификация — magic link и сессии ✅

**Цель:** Пользователь запрашивает magic link по email, переходит по ссылке, получает сессионную cookie. Middleware `RequireSession` защищает эндпоинты. В dev-режиме ссылка печатается в лог.

**Файлы:**
- Создать: `internal/auth/auth.go` — доменная логика: генерация токенов, валидация
- Создать: `internal/auth/auth_test.go`
- Создать: `internal/email/email.go` — интерфейс отправки + dev-заглушка + Resend-клиент
- Создать: `internal/email/email_test.go`
- Создать: `internal/http/middleware/session.go` — middleware RequireSession
- Создать: `internal/http/middleware/session_test.go`
- Создать: `internal/http/handlers/auth.go` — HTTP-обработчики auth
- Создать: `internal/http/handlers/auth_test.go`
- Изменить: `cmd/server/main.go` — подключить роуты auth

**Шаги:**

- [x] **Шаг 1: Создать `internal/auth/auth.go`**

```go
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateToken создает криптостойкий токен из 32 байт в base64url.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}
```

- [x] **Шаг 2: Написать тест для `GenerateToken`**

Создать `internal/auth/auth_test.go`:

```go
package auth_test

import (
	"testing"

	"github.com/andreypisarev/secret-santa/internal/auth"
)

func TestGenerateToken(t *testing.T) {
	token1, err := auth.GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(token1) == 0 {
		t.Fatal("token is empty")
	}

	token2, err := auth.GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token1 == token2 {
		t.Error("two tokens should not be equal")
	}

	// 32 байта в base64url без паддинга = 43 символа
	if len(token1) != 43 {
		t.Errorf("token length = %d, want 43", len(token1))
	}
}
```

- [x] **Шаг 3: Запустить тест**

```bash
go test ./internal/auth/ -v
```

Ожидаемый результат: PASS.

- [x] **Шаг 4: Создать `internal/email/email.go`**

```go
package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// Sender отправляет email.
type Sender interface {
	Send(to, subject, html string) error
}

// LogSender печатает письмо в лог (dev-режим).
type LogSender struct{}

func (s *LogSender) Send(to, subject, html string) error {
	slog.Info("email (dev)", "to", to, "subject", subject, "html", html)
	return nil
}

// ResendSender отправляет через Resend API.
type ResendSender struct {
	APIKey string
	From   string
}

func (s *ResendSender) Send(to, subject, html string) error {
	payload := map[string]interface{}{
		"from":    s.From,
		"to":     []string{to},
		"subject": subject,
		"html":    html,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal email: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: status %d", resp.StatusCode)
	}
	return nil
}
```

- [x] **Шаг 5: Написать тест для LogSender**

Создать `internal/email/email_test.go`:

```go
package email_test

import (
	"testing"

	"github.com/andreypisarev/secret-santa/internal/email"
)

func TestLogSender_Send(t *testing.T) {
	s := &email.LogSender{}
	err := s.Send("test@example.com", "Subject", "<p>Body</p>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
```

- [x] **Шаг 6: Создать `internal/http/middleware/session.go`**

```go
package middleware

import (
	"context"
	"database/sql"
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
```

- [x] **Шаг 7: Написать тест для middleware**

Создать `internal/http/middleware/session_test.go`:

```go
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
```

- [x] **Шаг 8: Создать `internal/http/handlers/auth.go`**

```go
package handlers

import (
	"database/sql"
	"encoding/json"
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
		ExpiresAt: expiresAt.Format(time.DateTime),
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

	// Найти или создать пользователя
	user, err := h.Queries.GetUserByEmail(r.Context(), ml.Email)
	if err == sql.ErrNoRows {
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

	// Создать сессию
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
		ExpiresAt: expiresAt.Format(time.DateTime),
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
		MaxAge:   30 * 24 * 60 * 60, // 30 дней
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
```

**Заметка:** Типы из sqlc-генерации (параметры `CreateMagicLinkParams`, `CreateSessionParams` и т.д.) могут отличаться по именам полей. При компиляции сверься с `internal/db/sqlc/*.go` и поправь имена, если нужно.

- [x] **Шаг 9: Написать тесты для auth handlers**

Создать `internal/http/handlers/auth_test.go`:

```go
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

func TestFullAuthFlow(t *testing.T) {
	h := setupAuth(t)

	// 1. Запросить magic link
	body := strings.NewReader(`{"email":"user@example.com"}`)
	req := httptest.NewRequest("POST", "/api/auth/request-link", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.RequestLink(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("request-link: status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// 2. Достать токен из БД напрямую (в тестах у нас доступ)
	// Так как LogSender не сохраняет токен, нам нужен прямой доступ к БД.
	// Для этого теста получим токен из magic_links таблицы.
	// Это интеграционный тест — доступ к БД допустим.
	var token string
	row := h.Queries.(*sqlc.Queries) // sqlc.Queries уже содержит db
	// Вместо этого — используем сам Queries, но GetMagicLink требует токен.
	// Для полного flow-теста нужно вытащить токен иначе.
	// Решение: используем прямой SQL через db, или мокаем Email sender.
	_ = row
	_ = token

	// Упрощенный подход: тестируем verify с токеном, созданным вручную.
	// Полный flow тестируется в smoke-тесте (фича 8).
	t.Log("full flow test deferred to smoke tests")
}

func TestLogout(t *testing.T) {
	h := setupAuth(t)

	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	w := httptest.NewRecorder()

	h.Logout(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// Проверяем, что cookie удалена
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

	// 1. Request link
	body := strings.NewReader(`{"email":"flow@example.com"}`)
	req := httptest.NewRequest("POST", "/api/auth/request-link", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.RequestLink(w, req)

	if len(sender.Sent) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(sender.Sent))
	}

	// Извлечь токен из ссылки в письме
	html := sender.Sent[0].HTML
	prefix := h.Config.BaseURL + "/api/auth/verify?token="
	idx := strings.Index(html, prefix)
	if idx == -1 {
		t.Fatalf("link not found in email: %s", html)
	}
	tokenStart := idx + len(prefix)
	tokenEnd := strings.Index(html[tokenStart:], `"`)
	token := html[tokenStart : tokenStart+tokenEnd]

	// 2. Verify
	req = httptest.NewRequest("GET", "/api/auth/verify?token="+token, nil)
	w = httptest.NewRecorder()
	h.Verify(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("verify: status = %d, want %d", w.Code, http.StatusFound)
	}

	// Проверяем cookie
	var sessionCookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "s" {
			sessionCookie = c
		}
	}
	if sessionCookie == nil {
		t.Fatal("session cookie not set")
	}

	// 3. Me — с cookie
	req = httptest.NewRequest("GET", "/api/auth/me", nil)
	req.AddCookie(sessionCookie)
	w = httptest.NewRecorder()

	// Нужно пройти через middleware
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
```

**Заметка:** Тест `TestVerify_FullFlow` использует хелпер `handlers.WithSession` — его нужно добавить в `auth.go`:

```go
// WithSession оборачивает handler в RequireSession middleware (удобно для тестов).
func WithSession(queries *sqlc.Queries, next http.Handler) http.Handler {
	return mw.RequireSession(queries)(next)
}
```

- [x] **Шаг 10: Подключить роуты auth в `cmd/server/main.go`**

```go
queries := sqlc.New(database)

var emailSender email.Sender
if cfg.IsDev() {
    emailSender = &email.LogSender{}
} else {
    emailSender = &email.ResendSender{
        APIKey: cfg.ResendAPIKey,
        From:   cfg.EmailFrom,
    }
}

authHandler := &handlers.AuthHandler{
    Queries: queries,
    Email:   emailSender,
    Config:  cfg,
}

r.Post("/api/auth/request-link", authHandler.RequestLink)
r.Get("/api/auth/verify", authHandler.Verify)
r.Post("/api/auth/logout", authHandler.Logout)

r.Group(func(r chi.Router) {
    r.Use(mw.RequireSession(queries))
    r.Get("/api/auth/me", authHandler.Me)
})
```

- [x] **Шаг 11: Запустить все тесты**

```bash
go test ./internal/... -v
```

Ожидаемый результат: все тесты PASS.

- [x] **Шаг 12: Коммит**

```bash
git add internal/auth/ internal/email/ internal/http/ cmd/server/main.go
git commit -m "feat: аутентификация — magic link, сессии, middleware"
```

**Проверка:**
- [x] `go test ./internal/... -v` — все тесты проходят
- [x] `go build ./cmd/server/` — компилируется

---

### Фича 4: Группы — создание, приглашение, вступление ✅

**Цель:** Организатор создает группу, получает invite-ссылку. Участник переходит по ссылке, видит название группы, вступает с именем и вишлистом. Страница группы показывает список участников.

**Файлы:**
- Создать: `internal/groups/groups.go` — генерация invite-кода
- Создать: `internal/groups/groups_test.go`
- Создать: `internal/http/handlers/groups.go` — HTTP-обработчики
- Создать: `internal/http/handlers/groups_test.go`
- Изменить: `cmd/server/main.go` — подключить роуты групп

**Шаги:**

- [x] **Шаг 1: Создать `internal/groups/groups.go`**

```go
package groups

import (
	"crypto/rand"
	"math/big"
)

const inviteCodeAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
const inviteCodeLength = 12

// GenerateInviteCode создает случайный код из 12 символов [a-z0-9].
func GenerateInviteCode() (string, error) {
	code := make([]byte, inviteCodeLength)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(inviteCodeAlphabet))))
		if err != nil {
			return "", err
		}
		code[i] = inviteCodeAlphabet[n.Int64()]
	}
	return string(code), nil
}
```

- [x] **Шаг 2: Написать тест для invite-кода**

Создать `internal/groups/groups_test.go`:

```go
package groups_test

import (
	"regexp"
	"testing"

	"github.com/andreypisarev/secret-santa/internal/groups"
)

func TestGenerateInviteCode(t *testing.T) {
	code, err := groups.GenerateInviteCode()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(code) != 12 {
		t.Errorf("length = %d, want 12", len(code))
	}

	if !regexp.MustCompile(`^[a-z0-9]{12}$`).MatchString(code) {
		t.Errorf("code %q doesn't match [a-z0-9]{12}", code)
	}

	// Два кода не должны совпадать
	code2, _ := groups.GenerateInviteCode()
	if code == code2 {
		t.Error("two codes should not be equal")
	}
}
```

- [x] **Шаг 3: Запустить тест**

```bash
go test ./internal/groups/ -v
```

Ожидаемый результат: PASS.

- [x] **Шаг 4: Создать `internal/http/handlers/groups.go`**

```go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"unicode/utf8"

	"github.com/andreypisarev/secret-santa/internal/db/sqlc"
	"github.com/andreypisarev/secret-santa/internal/groups"
	mw "github.com/andreypisarev/secret-santa/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

type GroupHandler struct {
	Queries *sqlc.Queries
}

func (h *GroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := mw.UserIDFromContext(r.Context())

	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "неверный формат запроса")
		return
	}

	if req.Title == "" || utf8.RuneCountInString(req.Title) > 100 {
		writeError(w, http.StatusBadRequest, "invalid_input", "название группы должно быть от 1 до 100 символов")
		return
	}

	code, err := groups.GenerateInviteCode()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	group, err := h.Queries.CreateGroup(r.Context(), sqlc.CreateGroupParams{
		InviteCode:  code,
		Title:       req.Title,
		OrganizerID: userID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":          group.ID,
		"invite_code": group.InviteCode,
	})
}

func (h *GroupHandler) GetByInviteCode(w http.ResponseWriter, r *http.Request) {
	inviteCode := chi.URLParam(r, "inviteCode")

	group, err := h.Queries.GetGroupByInviteCode(r.Context(), inviteCode)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not_found", "группа не найдена")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	userID, hasSession := mw.UserIDFromContext(r.Context())

	// Если нет сессии — минимальная информация
	if !hasSession {
		count, _ := h.Queries.CountMembersByGroup(r.Context(), group.ID)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"title":        group.Title,
			"member_count": count,
			"status":       group.Status,
		})
		return
	}

	// Если есть сессия — проверяем membership
	members, err := h.Queries.ListMembershipsByGroup(r.Context(), group.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	isMember := false
	isOrganizer := group.OrganizerID == userID
	var myMembershipID *int64

	memberList := make([]map[string]interface{}, 0, len(members))
	for _, m := range members {
		user, _ := h.Queries.GetUserByID(r.Context(), m.UserID)
		memberList = append(memberList, map[string]interface{}{
			"name":  user.Name,
			"is_me": m.UserID == userID,
		})
		if m.UserID == userID {
			isMember = true
			id := m.ID
			myMembershipID = &id
		}
	}

	if !isMember {
		count, _ := h.Queries.CountMembersByGroup(r.Context(), group.ID)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"title":        group.Title,
			"member_count": count,
			"status":       group.Status,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"title":            group.Title,
		"status":           group.Status,
		"members":          memberList,
		"is_organizer":     isOrganizer,
		"my_membership_id": myMembershipID,
	})
}

func (h *GroupHandler) Join(w http.ResponseWriter, r *http.Request) {
	userID, _ := mw.UserIDFromContext(r.Context())
	inviteCode := chi.URLParam(r, "inviteCode")

	var req struct {
		Name     string `json:"name"`
		Wishlist string `json:"wishlist"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "неверный формат запроса")
		return
	}

	if req.Name == "" || utf8.RuneCountInString(req.Name) > 50 {
		writeError(w, http.StatusBadRequest, "invalid_input", "имя должно быть от 1 до 50 символов")
		return
	}
	if utf8.RuneCountInString(req.Wishlist) > 2000 {
		writeError(w, http.StatusBadRequest, "invalid_input", "вишлист не должен превышать 2000 символов")
		return
	}

	group, err := h.Queries.GetGroupByInviteCode(r.Context(), inviteCode)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not_found", "группа не найдена")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	if group.Status != "open" {
		writeError(w, http.StatusConflict, "already_drawn", "жеребьевка уже проведена")
		return
	}

	// Обновить имя пользователя
	_ = h.Queries.UpdateUserName(r.Context(), sqlc.UpdateUserNameParams{
		Name: req.Name,
		ID:   userID,
	})

	_, err = h.Queries.CreateMembership(r.Context(), sqlc.CreateMembershipParams{
		GroupID:  group.ID,
		UserID:   userID,
		Wishlist: req.Wishlist,
	})
	if err != nil {
		// Скорее всего UNIQUE constraint — уже участник
		writeError(w, http.StatusConflict, "already_member", "вы уже участник этой группы")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupHandler) UpdateWishlist(w http.ResponseWriter, r *http.Request) {
	userID, _ := mw.UserIDFromContext(r.Context())
	membershipID := chi.URLParam(r, "id")

	var req struct {
		Wishlist string `json:"wishlist"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "неверный формат запроса")
		return
	}

	if utf8.RuneCountInString(req.Wishlist) > 2000 {
		writeError(w, http.StatusBadRequest, "invalid_input", "вишлист не должен превышать 2000 символов")
		return
	}

	// Парсим ID из строки
	var id int64
	for _, c := range membershipID {
		id = id*10 + int64(c-'0')
	}

	membership, err := h.Queries.GetMembership(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "участие не найдено")
		return
	}

	if membership.UserID != userID {
		writeError(w, http.StatusForbidden, "forbidden", "нет доступа")
		return
	}

	if err := h.Queries.UpdateWishlist(r.Context(), sqlc.UpdateWishlistParams{
		Wishlist: req.Wishlist,
		ID:       id,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

**Заметка:** Парсинг `membershipID` — упрощенный. Лучше использовать `strconv.ParseInt`. Исполнитель должен заменить ручной парсинг на `strconv.ParseInt(membershipID, 10, 64)`.

- [x] **Шаг 5: Написать тесты для group handlers**

Создать `internal/http/handlers/groups_test.go`:

```go
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

	// Создать пользователя
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

	// Создать организатора и группу
	org, _ := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		Email: "org@example.com", Name: "Организатор",
	})
	group, _ := queries.CreateGroup(context.Background(), sqlc.CreateGroupParams{
		InviteCode: "testcode1234", Title: "Тест", OrganizerID: org.ID,
	})

	// Создать участника
	member, _ := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		Email: "member@example.com", Name: "",
	})

	// Вступить в группу
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
```

- [x] **Шаг 6: Подключить роуты групп в `cmd/server/main.go`**

```go
groupHandler := &handlers.GroupHandler{Queries: queries}

r.Group(func(r chi.Router) {
    r.Use(mw.RequireSession(queries))
    r.Post("/api/groups", groupHandler.Create)
    r.Post("/api/groups/{inviteCode}/join", groupHandler.Join)
    r.Patch("/api/memberships/{id}", groupHandler.UpdateWishlist)
})

r.Group(func(r chi.Router) {
    r.Use(mw.OptionalSession(queries))
    r.Get("/api/groups/{inviteCode}", groupHandler.GetByInviteCode)
})
```

- [x] **Шаг 7: Запустить все тесты**

```bash
go test ./internal/... -v
```

Ожидаемый результат: все тесты PASS.

- [x] **Шаг 8: Коммит**

```bash
git add internal/groups/ internal/http/handlers/groups.go internal/http/handlers/groups_test.go cmd/server/main.go
git commit -m "feat: группы — создание, приглашение, вступление, обновление вишлиста"
```

**Проверка:**
- [x] `go test ./internal/... -v` — все тесты проходят
- [x] `go build ./cmd/server/` — компилируется

---

### Фича 5: Жеребьевка ✅

**Цель:** Организатор запускает жеребьевку. Алгоритм распределяет участников в цикл. Каждый может посмотреть своего подопечного.

**Файлы:**
- Создать: `internal/draw/draw.go` — алгоритм жеребьевки
- Создать: `internal/draw/draw_test.go`
- Создать: `internal/http/handlers/draw.go` — HTTP-обработчик
- Создать: `internal/http/handlers/draw_test.go`
- Изменить: `cmd/server/main.go` — подключить роуты

**Шаги:**

- [x] **Шаг 1: Создать `internal/draw/draw.go`**

```go
package draw

import (
	"errors"
	"math/rand"
)

var ErrNotEnoughMembers = errors.New("not_enough_members")

// Assign распределяет участников в один цикл.
// Каждый дарит следующему: perm[i] → perm[(i+1) % n].
// Возвращает map[sanтaID]recipientID.
func Assign(participants []int64, rng *rand.Rand) (map[int64]int64, error) {
	n := len(participants)
	if n < 2 {
		return nil, ErrNotEnoughMembers
	}

	// Копируем, чтобы не менять оригинал
	perm := make([]int64, n)
	copy(perm, participants)

	rng.Shuffle(n, func(i, j int) {
		perm[i], perm[j] = perm[j], perm[i]
	})

	result := make(map[int64]int64, n)
	for i := 0; i < n; i++ {
		santa := perm[i]
		recipient := perm[(i+1)%n]
		result[santa] = recipient
	}

	return result, nil
}
```

- [x] **Шаг 2: Написать тесты для жеребьевки**

Создать `internal/draw/draw_test.go`:

```go
package draw_test

import (
	"math/rand"
	"testing"

	"github.com/andreypisarev/secret-santa/internal/draw"
)

func TestAssign_NotEnoughMembers(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	_, err := draw.Assign([]int64{}, rng)
	if err != draw.ErrNotEnoughMembers {
		t.Errorf("err = %v, want ErrNotEnoughMembers", err)
	}

	_, err = draw.Assign([]int64{1}, rng)
	if err != draw.ErrNotEnoughMembers {
		t.Errorf("err = %v, want ErrNotEnoughMembers", err)
	}
}

func TestAssign_TwoMembers(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	result, err := draw.Assign([]int64{1, 2}, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// С двумя участниками: каждый дарит другому
	if result[1] == 1 || result[2] == 2 {
		t.Error("someone is assigned to themselves")
	}
	if len(result) != 2 {
		t.Errorf("result size = %d, want 2", len(result))
	}
}

func TestAssign_Invariants(t *testing.T) {
	participants := []int64{10, 20, 30, 40, 50, 60}
	rng := rand.New(rand.NewSource(123))

	result, err := draw.Assign(participants, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Инвариант 1: никто не дарит себе
	for santa, recipient := range result {
		if santa == recipient {
			t.Errorf("santa %d assigned to self", santa)
		}
	}

	// Инвариант 2: все участники назначены как санты
	if len(result) != len(participants) {
		t.Errorf("result size = %d, want %d", len(result), len(participants))
	}

	// Инвариант 3: у каждого ровно один подопечный
	recipients := make(map[int64]bool)
	for _, r := range result {
		if recipients[r] {
			t.Errorf("recipient %d assigned twice", r)
		}
		recipients[r] = true
	}

	// Инвариант 4: один большой цикл
	visited := make(map[int64]bool)
	current := participants[0]
	for i := 0; i < len(participants); i++ {
		if visited[current] {
			t.Fatalf("cycle broken at step %d, node %d", i, current)
		}
		visited[current] = true
		current = result[current]
	}
	if current != participants[0] {
		t.Error("did not return to start — not a single cycle")
	}
	if len(visited) != len(participants) {
		t.Errorf("visited %d nodes, want %d", len(visited), len(participants))
	}
}

func TestAssign_Deterministic(t *testing.T) {
	participants := []int64{1, 2, 3, 4, 5}

	r1, _ := draw.Assign(participants, rand.New(rand.NewSource(999)))
	r2, _ := draw.Assign(participants, rand.New(rand.NewSource(999)))

	for k, v := range r1 {
		if r2[k] != v {
			t.Errorf("not deterministic: key %d, got %d vs %d", k, v, r2[k])
		}
	}
}

func TestAssign_DoesNotMutateInput(t *testing.T) {
	original := []int64{1, 2, 3, 4, 5}
	input := make([]int64, len(original))
	copy(input, original)

	draw.Assign(input, rand.New(rand.NewSource(42)))

	for i, v := range input {
		if v != original[i] {
			t.Errorf("input[%d] = %d, was %d — input was mutated", i, v, original[i])
		}
	}
}
```

- [x] **Шаг 3: Запустить тесты жеребьевки**

```bash
go test ./internal/draw/ -v
```

Ожидаемый результат: все 5 тестов PASS.

- [x] **Шаг 4: Создать `internal/http/handlers/draw.go`**

```go
package handlers

import (
	"database/sql"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/andreypisarev/secret-santa/internal/db/sqlc"
	"github.com/andreypisarev/secret-santa/internal/draw"
	mw "github.com/andreypisarev/secret-santa/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

type DrawHandler struct {
	Queries *sqlc.Queries
	DB      *sql.DB // для транзакций
}

func (h *DrawHandler) Draw(w http.ResponseWriter, r *http.Request) {
	userID, _ := mw.UserIDFromContext(r.Context())
	groupIDStr := chi.URLParam(r, "id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "неверный ID группы")
		return
	}

	group, err := h.Queries.GetGroupByID(r.Context(), groupID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not_found", "группа не найдена")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	if group.OrganizerID != userID {
		writeError(w, http.StatusForbidden, "forbidden", "только организатор может провести жеребьевку")
		return
	}

	if group.Status != "open" {
		writeError(w, http.StatusConflict, "already_drawn", "жеребьевка уже проведена")
		return
	}

	members, err := h.Queries.ListMembershipsByGroup(r.Context(), groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	participantIDs := make([]int64, len(members))
	for i, m := range members {
		participantIDs[i] = m.UserID
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	assignments, err := draw.Assign(participantIDs, rng)
	if err == draw.ErrNotEnoughMembers {
		writeError(w, http.StatusBadRequest, "not_enough_members", "нужно минимум 2 участника")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	// Одна транзакция: обновить recipient_id + status
	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}
	defer tx.Rollback()

	qtx := h.Queries.WithTx(tx)

	res, err := qtx.DrawGroup(r.Context(), groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		writeError(w, http.StatusConflict, "already_drawn", "жеребьевка уже проведена")
		return
	}

	for santaID, recipientID := range assignments {
		if err := qtx.SetRecipient(r.Context(), sqlc.SetRecipientParams{
			RecipientID: &recipientID,
			GroupID:     groupID,
			UserID:      santaID,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DrawHandler) MyRecipient(w http.ResponseWriter, r *http.Request) {
	userID, _ := mw.UserIDFromContext(r.Context())
	groupIDStr := chi.URLParam(r, "id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "неверный ID группы")
		return
	}

	recipient, err := h.Queries.GetMyRecipient(r.Context(), sqlc.GetMyRecipientParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "not_found", "подопечный не назначен")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"recipient": map[string]interface{}{
			"name":     recipient.Name,
			"wishlist": recipient.Wishlist,
		},
	})
}
```

**Заметка:** `h.Queries.WithTx(tx)` — метод, который sqlc генерирует автоматически. Параметр `SetRecipientParams.RecipientID` может быть `*int64` или `sql.NullInt64` — зависит от sqlc-генерации. Сверить при компиляции.

- [x] **Шаг 5: Написать тест для draw handler**

Создать `internal/http/handlers/draw_test.go`:

```go
package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/andreypisarev/secret-santa/internal/db"
	dbpkg "github.com/andreypisarev/secret-santa/internal/db"
	"github.com/andreypisarev/secret-santa/internal/db/sqlc"
	"github.com/andreypisarev/secret-santa/internal/http/handlers"
	"github.com/go-chi/chi/v5"
)

func TestDraw(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer database.Close()
	dbpkg.Migrate(database)

	queries := sqlc.New(database)
	h := &handlers.DrawHandler{Queries: queries, DB: database}

	// Создать организатора + 3 участника
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

	// Организатор тоже участник
	queries.CreateMembership(context.Background(), sqlc.CreateMembershipParams{
		GroupID: group.ID, UserID: org.ID, Wishlist: "Мой подарок",
	})

	// Провести жеребьевку
	r := chi.NewRouter()
	r.Post("/api/groups/{id}/draw", h.Draw)

	req := httptest.NewRequest("POST", "/api/groups/"+strconv.FormatInt(group.ID, 10)+"/draw", nil)
	req = withUserID(req, org.ID)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body: %s", w.Code, http.StatusNoContent, w.Body.String())
	}

	// Проверяем: группа в статусе drawn
	updatedGroup, _ := queries.GetGroupByID(context.Background(), group.ID)
	if updatedGroup.Status != "drawn" {
		t.Errorf("group status = %q, want %q", updatedGroup.Status, "drawn")
	}

	// Проверяем: у каждого есть подопечный
	members, _ := queries.ListMembershipsByGroup(context.Background(), group.ID)
	for _, m := range members {
		if m.RecipientID == nil {
			t.Errorf("member %d has no recipient", m.UserID)
		}
	}
}

func TestDraw_NotOrganizer(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer database.Close()
	db.Migrate(database)

	queries := sqlc.New(database)
	h := &handlers.DrawHandler{Queries: queries, DB: database}

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
```

- [x] **Шаг 6: Подключить роуты draw в `cmd/server/main.go`**

```go
drawHandler := &handlers.DrawHandler{Queries: queries, DB: database}

// Внутри группы с RequireSession:
r.Post("/api/groups/{id}/draw", drawHandler.Draw)
r.Get("/api/groups/{id}/my-recipient", drawHandler.MyRecipient)
```

- [x] **Шаг 7: Запустить все тесты**

```bash
go test ./internal/... -v
```

Ожидаемый результат: все тесты PASS.

- [x] **Шаг 8: Коммит**

```bash
git add internal/draw/ internal/http/handlers/draw.go internal/http/handlers/draw_test.go cmd/server/main.go
git commit -m "feat: жеребьевка — алгоритм цикла, транзакция, проверка прав"
```

**Проверка:**
- [x] `go test ./internal/draw/ -v` — тесты алгоритма проходят
- [x] `go test ./internal/... -v` — все тесты проходят

---

### Фича 6: Чат — WebSocket-хаб и сообщения ✅

**Цель:** После жеребьевки участники могут анонимно переписываться через WebSocket. У каждого два чата: «как Санта» и «как подопечный». REST-эндпоинт отдает историю.

**Файлы:**
- Создать: `internal/chat/hub.go` — Hub-per-group, маршрутизация сообщений
- Создать: `internal/chat/client.go` — WebSocket-клиент
- Создать: `internal/chat/hub_test.go`
- Создать: `internal/http/handlers/chat.go` — WebSocket handler + REST история
- Создать: `internal/http/handlers/chat_test.go`
- Изменить: `cmd/server/main.go` — подключить роуты

**Шаги:**

- [x] **Шаг 1: Создать `internal/chat/client.go`**

```go
package chat

import (
	"context"
	"encoding/json"

	"github.com/coder/websocket"
)

type Client struct {
	UserID  int64
	GroupID int64
	Conn    *websocket.Conn
	Send    chan []byte
}

type InboundMessage struct {
	Type string `json:"type"` // "send"
	Role string `json:"role"` // "santa" | "recipient"
	Body string `json:"body"`
}

type OutboundMessage struct {
	Type      string `json:"type"`       // "message" | "drawn" | "error"
	ID        int64  `json:"id,omitempty"`
	Role      string `json:"role,omitempty"`
	FromMe    bool   `json:"from_me,omitempty"`
	Body      string `json:"body,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

func (c *Client) WritePump(ctx context.Context) {
	defer c.Conn.CloseNow()
	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				return
			}
			if err := c.Conn.Write(ctx, websocket.MessageText, msg); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) ReadPump(ctx context.Context, incoming chan<- ClientMessage) {
	defer c.Conn.CloseNow()
	for {
		_, data, err := c.Conn.Read(ctx)
		if err != nil {
			return
		}
		var msg InboundMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		incoming <- ClientMessage{Client: c, Message: msg}
	}
}

type ClientMessage struct {
	Client  *Client
	Message InboundMessage
}
```

- [x] **Шаг 2: Создать `internal/chat/hub.go`**

```go
package chat

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/andreypisarev/secret-santa/internal/db/sqlc"
)

// Membership хранит кэшированные данные об участнике.
type Membership struct {
	UserID      int64
	RecipientID int64 // мой подопечный
	SantaID     int64 // мой Санта
}

type Hub struct {
	groupID     int64
	clients     map[int64]map[*Client]struct{} // userID → connections
	memberships map[int64]*Membership          // userID → membership
	register    chan *Client
	unregister  chan *Client
	incoming    chan ClientMessage
	quit        chan struct{}
	queries     *sqlc.Queries

	// Rate limiting: userID → последние отправки
	rateMu    sync.Mutex
	rateCount map[int64][]time.Time
}

func NewHub(groupID int64, queries *sqlc.Queries, memberships map[int64]*Membership) *Hub {
	return &Hub{
		groupID:     groupID,
		clients:     make(map[int64]map[*Client]struct{}),
		memberships: memberships,
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		incoming:    make(chan ClientMessage, 256),
		quit:        make(chan struct{}),
		queries:     queries,
		rateCount:   make(map[int64][]time.Time),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]struct{})
			}
			h.clients[client.UserID][client] = struct{}{}

		case client := <-h.unregister:
			if conns, ok := h.clients[client.UserID]; ok {
				delete(conns, client)
				close(client.Send)
				if len(conns) == 0 {
					delete(h.clients, client.UserID)
				}
			}

		case cm := <-h.incoming:
			h.handleMessage(cm)

		case <-h.quit:
			for uid, conns := range h.clients {
				for c := range conns {
					close(c.Send)
				}
				delete(h.clients, uid)
			}
			return
		}
	}
}

func (h *Hub) Stop() {
	close(h.quit)
}

func (h *Hub) Register(c *Client) {
	h.register <- c
}

func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

func (h *Hub) Incoming() chan<- ClientMessage {
	return h.incoming
}

func (h *Hub) handleMessage(cm ClientMessage) {
	sender := cm.Client
	msg := cm.Message

	if msg.Type != "send" {
		h.sendError(sender, "неизвестный тип сообщения")
		return
	}

	if msg.Role != "santa" && msg.Role != "recipient" {
		h.sendError(sender, "неверная роль")
		return
	}

	if msg.Body == "" || utf8.RuneCountInString(msg.Body) > 2000 {
		h.sendError(sender, "сообщение должно быть от 1 до 2000 символов")
		return
	}

	// Rate limit: 10 сообщений в минуту
	if !h.checkRate(sender.UserID) {
		h.sendError(sender, "слишком много сообщений, подождите")
		return
	}

	membership, ok := h.memberships[sender.UserID]
	if !ok {
		h.sendError(sender, "вы не участник группы")
		return
	}

	var dbSenderID, dbRecipientID int64
	var direction string

	if msg.Role == "santa" {
		// Пишу подопечному как Санта
		dbSenderID = sender.UserID
		dbRecipientID = membership.RecipientID
		direction = "santa_to_recipient"
	} else {
		// Пишу Санте как подопечный
		dbSenderID = sender.UserID
		dbRecipientID = membership.SantaID
		direction = "recipient_to_santa"
	}

	// Сохранить в БД
	saved, err := h.queries.CreateMessage(context.Background(), sqlc.CreateMessageParams{
		GroupID:     h.groupID,
		SenderID:    dbSenderID,
		RecipientID: dbRecipientID,
		Direction:   direction,
		Body:        msg.Body,
	})
	if err != nil {
		slog.Error("save message", "error", err)
		h.sendError(sender, "ошибка сохранения")
		return
	}

	// Отправить обоим участникам диалога
	// Для отправителя: role = msg.Role, from_me = true
	h.sendToUser(sender.UserID, OutboundMessage{
		Type:      "message",
		ID:        saved.ID,
		Role:      msg.Role,
		FromMe:    true,
		Body:      msg.Body,
		CreatedAt: saved.CreatedAt,
	})

	// Для получателя: вычислить role с его точки зрения
	var recipientRole string
	if direction == "santa_to_recipient" {
		recipientRole = "recipient" // получатель видит это в чате «от моего Санты»
	} else {
		recipientRole = "santa" // Санта видит это в чате «от моего подопечного»
	}

	h.sendToUser(dbRecipientID, OutboundMessage{
		Type:      "message",
		ID:        saved.ID,
		Role:      recipientRole,
		FromMe:    false,
		Body:      msg.Body,
		CreatedAt: saved.CreatedAt,
	})
}

func (h *Hub) BroadcastDrawn() {
	data, _ := json.Marshal(OutboundMessage{Type: "drawn"})
	for _, conns := range h.clients {
		for c := range conns {
			select {
			case c.Send <- data:
			default:
			}
		}
	}
}

func (h *Hub) sendToUser(userID int64, msg OutboundMessage) {
	data, _ := json.Marshal(msg)
	if conns, ok := h.clients[userID]; ok {
		for c := range conns {
			select {
			case c.Send <- data:
			default:
			}
		}
	}
}

func (h *Hub) sendError(c *Client, reason string) {
	data, _ := json.Marshal(OutboundMessage{Type: "error", Reason: reason})
	select {
	case c.Send <- data:
	default:
	}
}

func (h *Hub) checkRate(userID int64) bool {
	h.rateMu.Lock()
	defer h.rateMu.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)

	// Удалить старые записи
	filtered := h.rateCount[userID][:0]
	for _, t := range h.rateCount[userID] {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= 10 {
		h.rateCount[userID] = filtered
		return false
	}

	h.rateCount[userID] = append(filtered, now)
	return true
}

// HubManager управляет хабами для разных групп.
type HubManager struct {
	mu      sync.Mutex
	hubs    map[int64]*Hub
	queries *sqlc.Queries
	db      *sql.DB
}

func NewHubManager(queries *sqlc.Queries, db *sql.DB) *HubManager {
	return &HubManager{
		hubs:    make(map[int64]*Hub),
		queries: queries,
		db:      db,
	}
}

func (m *HubManager) GetOrCreateHub(groupID int64) (*Hub, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if hub, ok := m.hubs[groupID]; ok {
		return hub, nil
	}

	// Загрузить memberships для группы
	members, err := m.queries.ListMembershipsByGroup(context.Background(), groupID)
	if err != nil {
		return nil, err
	}

	// Построить карту: кто чей Санта
	recipientOf := make(map[int64]int64) // userID → recipientID
	for _, mem := range members {
		if mem.RecipientID != nil {
			recipientOf[mem.UserID] = *mem.RecipientID
		}
	}

	// Построить обратную карту: кто мой Санта
	santaOf := make(map[int64]int64) // recipientID → santaID
	for santa, recipient := range recipientOf {
		santaOf[recipient] = santa
	}

	memberships := make(map[int64]*Membership)
	for _, mem := range members {
		m := &Membership{
			UserID:      mem.UserID,
			RecipientID: recipientOf[mem.UserID],
			SantaID:     santaOf[mem.UserID],
		}
		memberships[mem.UserID] = m
	}

	hub := NewHub(groupID, m.queries, memberships)
	go hub.Run()
	m.hubs[groupID] = hub
	return hub, nil
}

func (m *HubManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, hub := range m.hubs {
		hub.Stop()
	}
}
```

**Заметка:** Тип `mem.RecipientID` зависит от sqlc-генерации — может быть `*int64` или `sql.NullInt64`. Подправить при компиляции.

- [x] **Шаг 3: Написать тесты для Hub**

Создать `internal/chat/hub_test.go`:

```go
package chat_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/andreypisarev/secret-santa/internal/chat"
	"github.com/andreypisarev/secret-santa/internal/db"
	"github.com/andreypisarev/secret-santa/internal/db/sqlc"
)

func setupChatTest(t *testing.T) (*sqlc.Queries, int64, map[int64]*chat.Membership) {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	db.Migrate(database)

	queries := sqlc.New(database)

	// Создать пользователей
	u1, _ := queries.CreateUser(context.Background(), sqlc.CreateUserParams{Email: "a@test.com", Name: "Alice"})
	u2, _ := queries.CreateUser(context.Background(), sqlc.CreateUserParams{Email: "b@test.com", Name: "Bob"})
	u3, _ := queries.CreateUser(context.Background(), sqlc.CreateUserParams{Email: "c@test.com", Name: "Carol"})

	// Создать группу
	group, _ := queries.CreateGroup(context.Background(), sqlc.CreateGroupParams{
		InviteCode: "chattest1234", Title: "Chat Test", OrganizerID: u1.ID,
	})

	// Цикл: Alice → Bob → Carol → Alice
	memberships := map[int64]*chat.Membership{
		u1.ID: {UserID: u1.ID, RecipientID: u2.ID, SantaID: u3.ID},
		u2.ID: {UserID: u2.ID, RecipientID: u3.ID, SantaID: u1.ID},
		u3.ID: {UserID: u3.ID, RecipientID: u1.ID, SantaID: u2.ID},
	}

	return queries, group.ID, memberships
}

func TestHub_MessageDelivery(t *testing.T) {
	queries, groupID, memberships := setupChatTest(t)

	hub := chat.NewHub(groupID, queries, memberships)
	go hub.Run()
	defer hub.Stop()

	// Alice (Санта для Bob)
	aliceClient := &chat.Client{
		UserID:  1, // Alice
		GroupID: groupID,
		Send:    make(chan []byte, 10),
	}
	// Bob (подопечный Alice)
	bobClient := &chat.Client{
		UserID:  2, // Bob
		GroupID: groupID,
		Send:    make(chan []byte, 10),
	}

	hub.Register(aliceClient)
	hub.Register(bobClient)

	// Alice пишет как Санта
	hub.Incoming() <- chat.ClientMessage{
		Client: aliceClient,
		Message: chat.InboundMessage{
			Type: "send",
			Role: "santa",
			Body: "Привет, подопечный!",
		},
	}

	// Alice должна получить подтверждение (from_me=true)
	select {
	case data := <-aliceClient.Send:
		var msg chat.OutboundMessage
		json.Unmarshal(data, &msg)
		if msg.Type != "message" || !msg.FromMe || msg.Role != "santa" {
			t.Errorf("alice got unexpected: %+v", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("alice timeout")
	}

	// Bob должен получить сообщение (from_me=false, role=recipient)
	select {
	case data := <-bobClient.Send:
		var msg chat.OutboundMessage
		json.Unmarshal(data, &msg)
		if msg.Type != "message" || msg.FromMe || msg.Role != "recipient" {
			t.Errorf("bob got unexpected: %+v", msg)
		}
		if msg.Body != "Привет, подопечный!" {
			t.Errorf("body = %q", msg.Body)
		}
	case <-time.After(time.Second):
		t.Fatal("bob timeout")
	}
}

func TestHub_RateLimit(t *testing.T) {
	queries, groupID, memberships := setupChatTest(t)

	hub := chat.NewHub(groupID, queries, memberships)
	go hub.Run()
	defer hub.Stop()

	client := &chat.Client{
		UserID:  1,
		GroupID: groupID,
		Send:    make(chan []byte, 20),
	}
	hub.Register(client)

	// Отправить 11 сообщений — 11-е должно быть отклонено
	for i := 0; i < 11; i++ {
		hub.Incoming() <- chat.ClientMessage{
			Client: client,
			Message: chat.InboundMessage{
				Type: "send",
				Role: "santa",
				Body: "msg",
			},
		}
	}

	// Подождать обработки
	time.Sleep(100 * time.Millisecond)

	// Собрать все сообщения
	errorCount := 0
	for {
		select {
		case data := <-client.Send:
			var msg chat.OutboundMessage
			json.Unmarshal(data, &msg)
			if msg.Type == "error" {
				errorCount++
			}
		default:
			goto done
		}
	}
done:
	if errorCount == 0 {
		t.Error("expected at least one rate limit error")
	}
}

func TestHub_BroadcastDrawn(t *testing.T) {
	queries, groupID, memberships := setupChatTest(t)

	hub := chat.NewHub(groupID, queries, memberships)
	go hub.Run()
	defer hub.Stop()

	c1 := &chat.Client{UserID: 1, GroupID: groupID, Send: make(chan []byte, 10)}
	c2 := &chat.Client{UserID: 2, GroupID: groupID, Send: make(chan []byte, 10)}
	hub.Register(c1)
	hub.Register(c2)

	time.Sleep(50 * time.Millisecond)
	hub.BroadcastDrawn()

	for _, c := range []*chat.Client{c1, c2} {
		select {
		case data := <-c.Send:
			var msg chat.OutboundMessage
			json.Unmarshal(data, &msg)
			if msg.Type != "drawn" {
				t.Errorf("expected drawn, got %s", msg.Type)
			}
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for drawn broadcast")
		}
	}
}
```

- [x] **Шаг 4: Запустить тесты чата**

```bash
go test ./internal/chat/ -v
```

Ожидаемый результат: все тесты PASS.

- [x] **Шаг 5: Создать `internal/http/handlers/chat.go`**

```go
package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/andreypisarev/secret-santa/internal/chat"
	"github.com/andreypisarev/secret-santa/internal/config"
	"github.com/andreypisarev/secret-santa/internal/db/sqlc"
	mw "github.com/andreypisarev/secret-santa/internal/http/middleware"
	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
)

type ChatHandler struct {
	Queries    *sqlc.Queries
	HubManager *chat.HubManager
	Config     *config.Config
}

func (h *ChatHandler) WebSocket(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	groupIDStr := chi.URLParam(r, "id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	// Проверить membership и статус группы
	group, err := h.Queries.GetGroupByID(r.Context(), groupID)
	if err != nil {
		http.Error(w, "group not found", http.StatusNotFound)
		return
	}
	if group.Status != "drawn" {
		http.Error(w, "group not drawn yet", http.StatusBadRequest)
		return
	}

	_, err = h.Queries.GetMembershipByGroupAndUser(r.Context(), sqlc.GetMembershipByGroupAndUserParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		http.Error(w, "not a member", http.StatusForbidden)
		return
	}

	// Проверить Origin
	acceptOptions := &websocket.AcceptOptions{}
	if !h.Config.IsDev() {
		acceptOptions.OriginPatterns = []string{h.Config.BaseURL}
	} else {
		acceptOptions.InsecureSkipVerify = true
	}

	conn, err := websocket.Accept(w, r, acceptOptions)
	if err != nil {
		slog.Error("websocket accept", "error", err)
		return
	}

	hub, err := h.HubManager.GetOrCreateHub(groupID)
	if err != nil {
		slog.Error("get hub", "error", err)
		conn.Close(websocket.StatusInternalError, "internal error")
		return
	}

	client := &chat.Client{
		UserID:  userID,
		GroupID: groupID,
		Conn:    conn,
		Send:    make(chan []byte, 256),
	}

	hub.Register(client)

	ctx := r.Context()
	go client.WritePump(ctx)
	client.ReadPump(ctx, hub.Incoming())
	hub.Unregister(client)
}

func (h *ChatHandler) History(w http.ResponseWriter, r *http.Request) {
	userID, _ := mw.UserIDFromContext(r.Context())
	groupIDStr := chi.URLParam(r, "id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "неверный ID группы")
		return
	}

	role := chi.URLParam(r, "role")
	if role != "santa" && role != "recipient" {
		writeError(w, http.StatusBadRequest, "invalid_input", "роль должна быть santa или recipient")
		return
	}

	// Определить пару диалога
	membership, err := h.Queries.GetMembershipByGroupAndUser(r.Context(), sqlc.GetMembershipByGroupAndUserParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		writeError(w, http.StatusForbidden, "forbidden", "вы не участник группы")
		return
	}

	var senderID, recipientID int64
	var direction string

	if role == "santa" {
		// Мой чат как Санта с подопечным — santa_to_recipient
		if membership.RecipientID == nil {
			writeError(w, http.StatusNotFound, "not_found", "жеребьевка не проведена")
			return
		}
		senderID = userID
		recipientID = *membership.RecipientID
		direction = "santa_to_recipient"
	} else {
		// Мой чат как подопечный с Сантой — recipient_to_santa
		// Найти моего Санту (кто имеет меня как подопечного)
		members, _ := h.Queries.ListMembershipsByGroup(r.Context(), groupID)
		var santaID int64
		for _, m := range members {
			if m.RecipientID != nil && *m.RecipientID == userID {
				santaID = m.UserID
				break
			}
		}
		if santaID == 0 {
			writeError(w, http.StatusNotFound, "not_found", "Санта не найден")
			return
		}
		senderID = santaID
		recipientID = userID
		direction = "santa_to_recipient"
	}

	_ = senderID
	_ = recipientID
	_ = direction

	// Загрузить сообщения оба направления для пары
	// Для роли santa: sender=me, recipient=подопечный, direction оба
	// Упрощение: загружаем оба направления пары и фильтруем
	// Лучше: отдельный запрос, который берет все сообщения пары
	// Пока используем ListMessages с direction

	// TODO: этот эндпоинт нужно доработать — текущие sqlc-запросы
	// фильтруют по одному direction, а чат содержит оба.
	// Исполнитель должен добавить sqlc-запрос ListChatMessages,
	// который загружает все сообщения пары (оба direction).

	writeJSON(w, http.StatusOK, []interface{}{})
}
```

**Заметка:** REST-история чата требует доработки sqlc-запроса. Исполнитель должен добавить в `internal/db/queries/messages.sql`:

```sql
-- name: ListChatMessages :many
SELECT * FROM messages
WHERE group_id = ?
  AND ((sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?))
ORDER BY created_at DESC
LIMIT 50;
```

И переписать `History` хэндлер, чтобы использовать его.

- [x] **Шаг 6: Подключить роуты чата в `cmd/server/main.go`**

```go
hubManager := chat.NewHubManager(queries, database)
defer hubManager.CloseAll()

chatHandler := &handlers.ChatHandler{
    Queries:    queries,
    HubManager: hubManager,
    Config:     cfg,
}

// WebSocket не проходит через chi middleware стандартно,
// но нам нужна проверка сессии.
// Решение: проверяем cookie внутри handler.
r.Get("/ws/groups/{id}", chatHandler.WebSocket)

r.Group(func(r chi.Router) {
    r.Use(mw.RequireSession(queries))
    r.Get("/api/groups/{id}/chats/{role}", chatHandler.History)
})
```

Также добавить WebSocket middleware для проверки сессии — передать через `OptionalSession` или проверять вручную в хэндлере (уже реализовано внутри `ChatHandler.WebSocket`).

- [x] **Шаг 7: Установить зависимость coder/websocket**

```bash
go get github.com/coder/websocket
```

- [x] **Шаг 8: Запустить все тесты**

```bash
go test ./internal/... -v
```

Ожидаемый результат: все тесты PASS.

- [x] **Шаг 9: Коммит**

```bash
git add internal/chat/ internal/http/handlers/chat.go internal/http/handlers/chat_test.go cmd/server/main.go go.mod go.sum
git commit -m "feat: чат — WebSocket-хаб, маршрутизация сообщений, rate limiting"
```

**Проверка:**
- [x] `go test ./internal/chat/ -v` — тесты хаба проходят
- [x] `go test ./internal/... -v` — все тесты проходят
- [x] `go build ./cmd/server/` — компилируется

---

### Фича 7: Фронтенд — роутинг, страницы, API-клиент ✅

**Цель:** React-приложение с роутингом: страница входа, страница группы (до и после жеребьевки), чат. Подключение к бэкенду через fetch и WebSocket.

**Файлы:**
- Создать: `web/src/api/client.ts` — обертка над fetch
- Создать: `web/src/api/ws.ts` — WebSocket-клиент с реконнектом
- Создать: `web/src/pages/LoginPage.tsx`
- Создать: `web/src/pages/GroupPage.tsx`
- Создать: `web/src/pages/CreateGroupPage.tsx`
- Создать: `web/src/pages/JoinPage.tsx`
- Создать: `web/src/components/ChatPanel.tsx`
- Создать: `web/src/components/MemberList.tsx`
- Создать: `web/src/components/RecipientCard.tsx`
- Изменить: `web/src/App.tsx` — роутинг
- Изменить: `web/src/main.tsx`

**Шаги:**

- [x] **Шаг 1: Установить react-router**

```bash
cd /Users/andreypisarev/other/secret-santa/web
npm install react-router
```

- [x] **Шаг 2: Создать `web/src/api/client.ts`**

```ts
interface ApiError {
  error: string;
  message: string;
}

class ApiClient {
  private async request<T>(url: string, options?: RequestInit): Promise<T> {
    const res = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options?.headers,
      },
    });

    if (!res.ok) {
      const body: ApiError = await res.json().catch(() => ({
        error: "unknown",
        message: "Неизвестная ошибка",
      }));
      throw body;
    }

    if (res.status === 204) {
      return undefined as T;
    }

    return res.json();
  }

  // Auth
  requestLink(email: string) {
    return this.request<void>("/api/auth/request-link", {
      method: "POST",
      body: JSON.stringify({ email }),
    });
  }

  getMe() {
    return this.request<{ user_id: number; email: string; name: string }>(
      "/api/auth/me"
    );
  }

  logout() {
    return this.request<void>("/api/auth/logout", { method: "POST" });
  }

  // Groups
  createGroup(title: string) {
    return this.request<{ id: number; invite_code: string }>("/api/groups", {
      method: "POST",
      body: JSON.stringify({ title }),
    });
  }

  getGroup(inviteCode: string) {
    return this.request<{
      title: string;
      status: string;
      member_count?: number;
      members?: { name: string; is_me: boolean }[];
      is_organizer?: boolean;
      my_membership_id?: number;
    }>(`/api/groups/${inviteCode}`);
  }

  joinGroup(inviteCode: string, name: string, wishlist: string) {
    return this.request<void>(`/api/groups/${inviteCode}/join`, {
      method: "POST",
      body: JSON.stringify({ name, wishlist }),
    });
  }

  updateWishlist(membershipId: number, wishlist: string) {
    return this.request<void>(`/api/memberships/${membershipId}`, {
      method: "PATCH",
      body: JSON.stringify({ wishlist }),
    });
  }

  // Draw
  draw(groupId: number) {
    return this.request<void>(`/api/groups/${groupId}/draw`, {
      method: "POST",
    });
  }

  getMyRecipient(groupId: number) {
    return this.request<{
      recipient: { name: string; wishlist: string };
    }>(`/api/groups/${groupId}/my-recipient`);
  }

  // Chat
  getChatHistory(groupId: number, role: "santa" | "recipient") {
    return this.request<
      { id: number; from_me: boolean; body: string; created_at: string }[]
    >(`/api/groups/${groupId}/chats/${role}`);
  }
}

export const api = new ApiClient();
```

- [x] **Шаг 3: Создать `web/src/api/ws.ts`**

```ts
type MessageHandler = (msg: {
  type: string;
  id?: number;
  role?: string;
  from_me?: boolean;
  body?: string;
  created_at?: string;
  reason?: string;
}) => void;

export class ChatSocket {
  private ws: WebSocket | null = null;
  private url: string;
  private onMessage: MessageHandler;
  private reconnectAttempts = 0;
  private maxReconnectDelay = 30000;
  private closed = false;

  constructor(groupId: number, onMessage: MessageHandler) {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    this.url = `${protocol}//${window.location.host}/ws/groups/${groupId}`;
    this.onMessage = onMessage;
    this.connect();
  }

  private connect() {
    if (this.closed) return;

    this.ws = new WebSocket(this.url);

    this.ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        this.onMessage(msg);
      } catch {
        // ignore parse errors
      }
    };

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
    };

    this.ws.onclose = () => {
      if (this.closed) return;
      this.scheduleReconnect();
    };

    this.ws.onerror = () => {
      this.ws?.close();
    };
  }

  private scheduleReconnect() {
    const delay = Math.min(
      1000 * Math.pow(2, this.reconnectAttempts) + Math.random() * 1000,
      this.maxReconnectDelay
    );
    this.reconnectAttempts++;
    setTimeout(() => this.connect(), delay);
  }

  send(role: "santa" | "recipient", body: string) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: "send", role, body }));
    }
  }

  close() {
    this.closed = true;
    this.ws?.close();
  }
}
```

- [x] **Шаг 4: Создать `web/src/pages/LoginPage.tsx`**

```tsx
import { useState } from "react";
import { api } from "../api/client";

export default function LoginPage({
  onLogin,
}: {
  onLogin: () => void;
}) {
  const [email, setEmail] = useState("");
  const [sent, setSent] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    try {
      await api.requestLink(email);
      setSent(true);
    } catch {
      setError("Не удалось отправить ссылку");
    }
  };

  if (sent) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="bg-white rounded-lg shadow p-8 max-w-md w-full text-center">
          <h1 className="text-2xl font-bold mb-4">Проверь почту</h1>
          <p className="text-gray-600">
            Мы отправили ссылку для входа на <strong>{email}</strong>
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow p-8 max-w-md w-full">
        <h1 className="text-2xl font-bold mb-6 text-center">Тайный Санта</h1>
        <form onSubmit={handleSubmit}>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Email
          </label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full border rounded-lg px-3 py-2 mb-4"
            placeholder="you@example.com"
            required
          />
          {error && <p className="text-red-500 text-sm mb-4">{error}</p>}
          <button
            type="submit"
            className="w-full bg-red-600 text-white rounded-lg py-2 font-medium hover:bg-red-700"
          >
            Получить ссылку для входа
          </button>
        </form>
      </div>
    </div>
  );
}
```

- [x] **Шаг 5: Создать `web/src/pages/CreateGroupPage.tsx`**

```tsx
import { useState } from "react";
import { useNavigate } from "react-router";
import { api } from "../api/client";

export default function CreateGroupPage() {
  const [title, setTitle] = useState("");
  const [error, setError] = useState("");
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    try {
      const group = await api.createGroup(title);
      navigate(`/g/${group.invite_code}`);
    } catch {
      setError("Не удалось создать группу");
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow p-8 max-w-md w-full">
        <h1 className="text-2xl font-bold mb-6">Создать группу</h1>
        <form onSubmit={handleSubmit}>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Название
          </label>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="w-full border rounded-lg px-3 py-2 mb-4"
            placeholder="Новый год 2026"
            maxLength={100}
            required
          />
          {error && <p className="text-red-500 text-sm mb-4">{error}</p>}
          <button
            type="submit"
            className="w-full bg-red-600 text-white rounded-lg py-2 font-medium hover:bg-red-700"
          >
            Создать
          </button>
        </form>
      </div>
    </div>
  );
}
```

- [x] **Шаг 6: Создать `web/src/components/MemberList.tsx`**

```tsx
interface Member {
  name: string;
  is_me: boolean;
}

export default function MemberList({ members }: { members: Member[] }) {
  return (
    <div className="bg-white rounded-lg shadow p-4">
      <h2 className="font-bold text-lg mb-3">Участники ({members.length})</h2>
      <ul className="space-y-1">
        {members.map((m, i) => (
          <li key={i} className="flex items-center gap-2">
            <span className="text-gray-800">{m.name}</span>
            {m.is_me && (
              <span className="text-xs bg-red-100 text-red-700 px-2 py-0.5 rounded">
                это ты
              </span>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}
```

- [x] **Шаг 7: Создать `web/src/components/RecipientCard.tsx`**

```tsx
export default function RecipientCard({
  name,
  wishlist,
}: {
  name: string;
  wishlist: string;
}) {
  return (
    <div className="bg-green-50 border border-green-200 rounded-lg p-4">
      <h2 className="font-bold text-lg mb-2">Твой подопечный: {name}</h2>
      {wishlist && (
        <div>
          <h3 className="font-medium text-sm text-gray-600 mb-1">Вишлист:</h3>
          <p className="text-gray-800 whitespace-pre-wrap">{wishlist}</p>
        </div>
      )}
    </div>
  );
}
```

- [x] **Шаг 8: Создать `web/src/components/ChatPanel.tsx`**

```tsx
import { useState, useEffect, useRef } from "react";
import { ChatSocket } from "../api/ws";
import { api } from "../api/client";

interface Message {
  id: number;
  from_me: boolean;
  body: string;
  created_at: string;
}

export default function ChatPanel({
  groupId,
  role,
  title,
}: {
  groupId: number;
  role: "santa" | "recipient";
  title: string;
}) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const socketRef = useRef<ChatSocket | null>(null);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // Загрузить историю
    api.getChatHistory(groupId, role).then((history) => {
      setMessages(history.reverse());
    });
  }, [groupId, role]);

  useEffect(() => {
    if (!socketRef.current) {
      socketRef.current = new ChatSocket(groupId, (msg) => {
        if (msg.type === "message" && msg.role === role) {
          setMessages((prev) => [
            ...prev,
            {
              id: msg.id!,
              from_me: msg.from_me!,
              body: msg.body!,
              created_at: msg.created_at!,
            },
          ]);
        }
      });
    }
    return () => {
      socketRef.current?.close();
      socketRef.current = null;
    };
  }, [groupId, role]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim()) return;
    socketRef.current?.send(role, input.trim());
    setInput("");
  };

  return (
    <div className="bg-white rounded-lg shadow flex flex-col h-80">
      <div className="px-4 py-2 border-b font-medium text-sm">{title}</div>
      <div className="flex-1 overflow-y-auto p-4 space-y-2">
        {messages.map((msg) => (
          <div
            key={msg.id}
            className={`flex ${msg.from_me ? "justify-end" : "justify-start"}`}
          >
            <div
              className={`rounded-lg px-3 py-2 max-w-xs ${
                msg.from_me
                  ? "bg-red-600 text-white"
                  : "bg-gray-100 text-gray-800"
              }`}
            >
              {msg.body}
            </div>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>
      <form onSubmit={handleSend} className="border-t p-2 flex gap-2">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          className="flex-1 border rounded-lg px-3 py-1"
          placeholder="Написать сообщение..."
          maxLength={2000}
        />
        <button
          type="submit"
          className="bg-red-600 text-white px-4 py-1 rounded-lg hover:bg-red-700"
        >
          Отправить
        </button>
      </form>
    </div>
  );
}
```

- [x] **Шаг 9: Создать `web/src/pages/JoinPage.tsx`**

```tsx
import { useState } from "react";
import { useNavigate, useParams } from "react-router";
import { api } from "../api/client";

export default function JoinPage() {
  const { inviteCode } = useParams<{ inviteCode: string }>();
  const [name, setName] = useState("");
  const [wishlist, setWishlist] = useState("");
  const [error, setError] = useState("");
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    try {
      await api.joinGroup(inviteCode!, name, wishlist);
      navigate(`/g/${inviteCode}`);
    } catch {
      setError("Не удалось вступить в группу");
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow p-8 max-w-md w-full">
        <h1 className="text-2xl font-bold mb-6">Присоединиться</h1>
        <form onSubmit={handleSubmit}>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Твое имя
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full border rounded-lg px-3 py-2 mb-4"
            maxLength={50}
            required
          />
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Вишлист (что хочешь получить)
          </label>
          <textarea
            value={wishlist}
            onChange={(e) => setWishlist(e.target.value)}
            className="w-full border rounded-lg px-3 py-2 mb-4 h-24"
            maxLength={2000}
          />
          {error && <p className="text-red-500 text-sm mb-4">{error}</p>}
          <button
            type="submit"
            className="w-full bg-red-600 text-white rounded-lg py-2 font-medium hover:bg-red-700"
          >
            Вступить
          </button>
        </form>
      </div>
    </div>
  );
}
```

- [x] **Шаг 10: Создать `web/src/pages/GroupPage.tsx`**

```tsx
import { useEffect, useState } from "react";
import { useParams } from "react-router";
import { api } from "../api/client";
import MemberList from "../components/MemberList";
import RecipientCard from "../components/RecipientCard";
import ChatPanel from "../components/ChatPanel";

interface GroupData {
  title: string;
  status: string;
  member_count?: number;
  members?: { name: string; is_me: boolean }[];
  is_organizer?: boolean;
  my_membership_id?: number;
}

export default function GroupPage({ userId }: { userId: number | null }) {
  const { inviteCode } = useParams<{ inviteCode: string }>();
  const [group, setGroup] = useState<GroupData | null>(null);
  const [recipient, setRecipient] = useState<{
    name: string;
    wishlist: string;
  } | null>(null);
  const [error, setError] = useState("");
  const [groupId, setGroupId] = useState<number | null>(null);

  const loadGroup = async () => {
    try {
      const data = await api.getGroup(inviteCode!);
      setGroup(data);
    } catch {
      setError("Группа не найдена");
    }
  };

  useEffect(() => {
    loadGroup();
  }, [inviteCode]);

  useEffect(() => {
    if (group?.status === "drawn" && groupId) {
      api.getMyRecipient(groupId).then((data) => {
        setRecipient(data.recipient);
      });
    }
  }, [group?.status, groupId]);

  const handleDraw = async () => {
    if (!groupId) return;
    try {
      await api.draw(groupId);
      loadGroup();
    } catch (e: any) {
      setError(e.message || "Не удалось провести жеребьевку");
    }
  };

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <p className="text-red-500">{error}</p>
      </div>
    );
  }

  if (!group) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <p className="text-gray-500">Загрузка...</p>
      </div>
    );
  }

  // Не участник — показать кнопку вступления
  if (!group.members) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="bg-white rounded-lg shadow p-8 max-w-md w-full text-center">
          <h1 className="text-2xl font-bold mb-2">{group.title}</h1>
          <p className="text-gray-600 mb-4">
            {group.member_count} участник(ов) | {group.status === "open" ? "Прием участников" : "Жеребьевка проведена"}
          </p>
          {group.status === "open" && userId && (
            <a
              href={`/g/${inviteCode}/join`}
              className="inline-block bg-red-600 text-white rounded-lg px-6 py-2 font-medium hover:bg-red-700"
            >
              Вступить
            </a>
          )}
          {!userId && (
            <p className="text-sm text-gray-500">
              Войдите, чтобы присоединиться
            </p>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 p-4">
      <div className="max-w-2xl mx-auto space-y-4">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">{group.title}</h1>
          <span
            className={`text-sm px-3 py-1 rounded-full ${
              group.status === "open"
                ? "bg-yellow-100 text-yellow-800"
                : "bg-green-100 text-green-800"
            }`}
          >
            {group.status === "open" ? "Прием участников" : "Жеребьевка проведена"}
          </span>
        </div>

        {/* Ссылка-приглашение */}
        {group.status === "open" && (
          <div className="bg-white rounded-lg shadow p-4">
            <p className="text-sm text-gray-600 mb-1">Ссылка для приглашения:</p>
            <code className="text-sm bg-gray-100 p-2 rounded block">
              {window.location.origin}/g/{inviteCode}
            </code>
          </div>
        )}

        <MemberList members={group.members} />

        {/* Кнопка жеребьевки для организатора */}
        {group.is_organizer && group.status === "open" && (
          <button
            onClick={handleDraw}
            className="w-full bg-green-600 text-white rounded-lg py-3 font-medium hover:bg-green-700"
          >
            Провести жеребьевку
          </button>
        )}

        {/* После жеребьевки */}
        {group.status === "drawn" && recipient && (
          <>
            <RecipientCard name={recipient.name} wishlist={recipient.wishlist} />
            {groupId && (
              <div className="space-y-4">
                <ChatPanel
                  groupId={groupId}
                  role="santa"
                  title="Переписка с подопечным (ты — Санта)"
                />
                <ChatPanel
                  groupId={groupId}
                  role="recipient"
                  title="Переписка с твоим Сантой"
                />
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
```

**Заметка:** `groupId` (числовой) не приходит из текущего API `GET /api/groups/:inviteCode`. Исполнитель должен либо добавить `id` в ответ API, либо использовать другой способ получения groupId. Простейшее решение — добавить `id` в ответ `GetByInviteCode`.

- [x] **Шаг 11: Обновить `web/src/App.tsx` с роутингом**

```tsx
import { createBrowserRouter, RouterProvider } from "react-router";
import { useState, useEffect } from "react";
import { api } from "./api/client";
import LoginPage from "./pages/LoginPage";
import CreateGroupPage from "./pages/CreateGroupPage";
import GroupPage from "./pages/GroupPage";
import JoinPage from "./pages/JoinPage";

function App() {
  const [user, setUser] = useState<{
    user_id: number;
    email: string;
    name: string;
  } | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .getMe()
      .then(setUser)
      .catch(() => setUser(null))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <p className="text-gray-500">Загрузка...</p>
      </div>
    );
  }

  const router = createBrowserRouter([
    {
      path: "/",
      element: user ? (
        <CreateGroupPage />
      ) : (
        <LoginPage onLogin={() => window.location.reload()} />
      ),
    },
    {
      path: "/g/:inviteCode",
      element: <GroupPage userId={user?.user_id ?? null} />,
    },
    {
      path: "/g/:inviteCode/join",
      element: user ? (
        <JoinPage />
      ) : (
        <LoginPage onLogin={() => window.location.reload()} />
      ),
    },
  ]);

  return <RouterProvider router={router} />;
}

export default App;
```

- [x] **Шаг 12: Проверить, что фронт собирается**

```bash
cd /Users/andreypisarev/other/secret-santa/web
npm run build
```

Ожидаемый результат: сборка без ошибок.

- [x] **Шаг 13: Коммит**

```bash
git add web/
git commit -m "feat: фронтенд — роутинг, страницы входа/группы/чата, API-клиент"
```

**Проверка:**
- [x] `cd web && npm run build` — фронт собирается без ошибок
- [x] `cd web && npx tsc --noEmit` — проверка типов без ошибок

---

### Фича 8: Раздача статики и интеграционный smoke-тест ✅

**Цель:** Go-сервер раздает собранный фронт из `embed.FS`. Smoke-тест проверяет полный flow: регистрация → создание группы → join → draw → отправка сообщения.

**Файлы:**
- Создать: `internal/http/static.go` — раздача embed.FS
- Изменить: `cmd/server/main.go` — подключить раздачу статики
- Создать: `internal/http/handlers/smoke_test.go` — интеграционный тест

**Шаги:**

- [x] **Шаг 1: Создать `internal/http/static.go`**

```go
package http

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// StaticHandler раздает файлы из embed.FS.
// Для SPA: если файл не найден — отдает index.html.
func StaticHandler(dist embed.FS, prefix string) http.Handler {
	sub, err := fs.Sub(dist, prefix)
	if err != nil {
		panic("invalid embed prefix: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// API и WS не обслуживаются тут
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/ws/") || r.URL.Path == "/healthz" {
			http.NotFound(w, r)
			return
		}

		// Попробовать отдать файл
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		if _, err := fs.Stat(sub, strings.TrimPrefix(path, "/")); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback — отдать index.html
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
```

- [x] **Шаг 2: Подключить раздачу статики в `cmd/server/main.go`**

В конце роутов, после всех API-маршрутов:

```go
import (
    secretsanta "github.com/andreypisarev/secret-santa"
    httputil "github.com/andreypisarev/secret-santa/internal/http"
)

// В конце настройки роутера:
r.Handle("/*", httputil.StaticHandler(secretsanta.WebDist, "web/dist"))
```

- [x] **Шаг 3: Написать smoke-тест**

Создать `internal/http/handlers/smoke_test.go`:

```go
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
	db.Migrate(database)

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

	// Request link
	body := strings.NewReader(`{"email":"` + emailAddr + `"}`)
	req := httptest.NewRequest("POST", "/api/auth/request-link", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("request-link: %d", w.Code)
	}

	// Извлечь токен из письма
	html := sender.Sent[len(sender.Sent)-1].HTML
	prefix := "http://localhost/api/auth/verify?token="
	idx := strings.Index(html, prefix)
	if idx == -1 {
		t.Fatal("token not in email")
	}
	start := idx + len(prefix)
	end := strings.Index(html[start:], `"`)
	token := html[start : start+end]

	// Verify
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
	for i, email := range []string{"alice@test.com", "bob@test.com"} {
		cookie := login(t, r, sender, email)
		names := []string{"Алиса", "Боб"}

		body = strings.NewReader(`{"name":"` + names[i] + `","wishlist":"Подарок"}`)
		req = httptest.NewRequest("POST", "/api/groups/"+inviteCode+"/join", body)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(cookie)
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("join %s: %d", email, w.Code)
		}
	}

	// 5. Организатор проводит жеребьевку
	req = httptest.NewRequest("POST", "/api/groups/"+strings.Itoa(groupID)+"/draw", nil)
	req.AddCookie(orgCookie)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("draw: %d, body: %s", w.Code, w.Body.String())
	}

	// 6. Организатор видит своего подопечного
	req = httptest.NewRequest("GET", "/api/groups/"+strings.Itoa(groupID)+"/my-recipient", nil)
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
```

- [x] **Шаг 4: Запустить smoke-тест**

```bash
go test ./internal/http/handlers/ -run TestSmoke -v
```

Ожидаемый результат: PASS.

- [x] **Шаг 5: Запустить все тесты**

```bash
go test ./internal/... -v
cd /Users/andreypisarev/other/secret-santa/web && npm run build && npx tsc --noEmit
```

Ожидаемый результат: все Go-тесты PASS, фронт собирается и проходит проверку типов.

- [x] **Шаг 6: Коммит**

```bash
git add internal/http/static.go internal/http/handlers/smoke_test.go cmd/server/main.go
git commit -m "feat: раздача статики + smoke-тест полного flow"
```

**Проверка:**
- [x] `go test ./internal/... -v` — все тесты проходят
- [x] `go build ./cmd/server/` — компилируется

---

### Фича 9: Деплой — Dockerfile и fly.toml ✅

**Цель:** Приложение собирается в Docker-образ и готово к деплою на Fly.io.

**Файлы:**
- Создать: `Dockerfile`
- Создать: `fly.toml`

**Шаги:**

- [x] **Шаг 1: Создать `Dockerfile`**

```dockerfile
# Stage 1: build frontend
FROM node:20-alpine AS web
WORKDIR /web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: build Go binary with embedded frontend
FROM golang:1.22-alpine AS server
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /web/dist ./web/dist
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

# Stage 3: minimal runtime
FROM alpine:3.20
RUN adduser -D -u 1000 app
COPY --from=server /out/server /usr/local/bin/server
USER app
EXPOSE 8080
CMD ["/usr/local/bin/server"]
```

- [x] **Шаг 2: Создать `fly.toml`**

```toml
app = "secret-santa"
primary_region = "ams"

[build]

[env]
  PORT = "8080"
  ENV = "production"
  LOG_LEVEL = "info"
  DATABASE_PATH = "/data/app.db"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = "stop"
  auto_start_machines = true
  min_machines_running = 0

[[vm]]
  size = "shared-cpu-1x"
  memory = "256mb"

[[mounts]]
  source = "data"
  destination = "/data"
```

- [x] **Шаг 3: Проверить Docker-сборку (если Docker установлен)**

```bash
cd /Users/andreypisarev/other/secret-santa
docker build -t secret-santa .
```

Ожидаемый результат: образ собран.

- [x] **Шаг 4: Коммит**

```bash
git add Dockerfile fly.toml
git commit -m "feat: Dockerfile + fly.toml для деплоя на Fly.io"
```

**Проверка:**
- [x] `docker build -t secret-santa .` — образ собирается (если Docker доступен)

---

### Фича 10: CI — GitHub Actions

**Цель:** При пуше запускаются Go-тесты, проверка типов фронта и сборка.

**Файлы:**
- Создать: `.github/workflows/ci.yml`

**Шаги:**

- [ ] **Шаг 1: Создать `.github/workflows/ci.yml`**

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - uses: actions/setup-node@v4
        with:
          node-version: "20"

      - name: Install frontend dependencies
        working-directory: web
        run: npm ci

      - name: Build frontend
        working-directory: web
        run: npm run build

      - name: TypeScript check
        working-directory: web
        run: npx tsc --noEmit

      - name: Go tests
        run: go test ./internal/... -v

      - name: Go build
        run: go build ./cmd/server/
```

- [ ] **Шаг 2: Коммит**

```bash
git add .github/
git commit -m "ci: GitHub Actions — тесты Go, typecheck и сборка фронта"
```

**Проверка:**
- [ ] Файл `.github/workflows/ci.yml` корректен по YAML-синтаксису
