# План 1 — Foundations

> **Для агентов:** REQUIRED SUB-SKILL: используйте superpowers:subagent-driven-development (рекомендуется) или superpowers:executing-plans для пошагового выполнения. Шаги помечены через `- [ ]`.

**Цель:** Собрать каркас проекта (Go-бэкенд + React/Vite-фронт) с работающим health-check, Dockerfile и dev-workflow. После этого плана локально запускается `make dev`, отвечает `GET /healthz`, открывается пустая страница от Vite с проксированием `/api/*` на Go.

**Архитектура:** Монолит: Go-сервер (`chi`) отдает API и статические файлы фронта (через `//go:embed`). В dev-режиме фронт запускается отдельно через Vite и проксирует запросы на Go. В проде все в одном бинарнике.

**Стек:** Go 1.22+, chi v5, slog, godotenv; React 18, Vite 5, TypeScript, Tailwind CSS v3, react-router v7.

**Контекст проекта:** сейчас в репозитории только `README.md` и `docs/project-design.md` (спецификация). Рабочая директория — `/Users/andreypisarev/other/secret-santa`. Все команды запускаются из корня репозитория, если не указано иное.

---

## Task 1: Инициализировать Go-модуль и каркас папок

**Files:**
- Create: `go.mod`
- Create: `.gitignore`
- Create: `cmd/server/.gitkeep`
- Create: `internal/.gitkeep`
- Create: `web/.gitkeep`

- [ ] **Шаг 1: Инициализировать Go-модуль**

```bash
go mod init secret-santa
```

Ожидаемый вывод: `go: creating new go.mod: module secret-santa`

- [ ] **Шаг 2: Создать структуру папок**

```bash
mkdir -p cmd/server internal/config internal/http/handlers internal/http/middleware web/src
```

- [ ] **Шаг 3: Создать `.gitignore`**

Создать файл `.gitignore` в корне:

```
# Go
*.exe
*.test
*.out
vendor/

# Node / Vite
node_modules/
web/dist/
.DS_Store

# Env
.env
.env.local

# Local DB
*.db
*.db-journal

# Editor
.vscode/
.idea/
*.swp
```

- [ ] **Шаг 4: Закоммитить каркас**

```bash
git add go.mod .gitignore
git commit -m "Инициализировать Go-модуль и каркас проекта"
```

---

## Task 2: Пакет `internal/config` (TDD)

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

Конфиг читается из переменных окружения. Значения по умолчанию позволяют запустить сервер в dev-режиме без `.env`.

- [ ] **Шаг 1: Написать тест**

Создать `internal/config/config_test.go`:

```go
package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("ENV", "")
	t.Setenv("PORT", "")
	t.Setenv("BASE_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Env != "development" {
		t.Errorf("Env = %q, want %q", cfg.Env, "development")
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}
	if cfg.BaseURL != "http://localhost:8080" {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, "http://localhost:8080")
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("ENV", "production")
	t.Setenv("PORT", "9090")
	t.Setenv("BASE_URL", "https://example.com")
	t.Setenv("DATABASE_PATH", "/data/app.db")
	t.Setenv("SESSION_SECRET", "test-secret")
	t.Setenv("RESEND_API_KEY", "re_xxx")
	t.Setenv("EMAIL_FROM", "noreply@example.com")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Env != "production" {
		t.Errorf("Env = %q, want %q", cfg.Env, "production")
	}
	if cfg.Port != "9090" {
		t.Errorf("Port = %q", cfg.Port)
	}
	if cfg.BaseURL != "https://example.com" {
		t.Errorf("BaseURL = %q", cfg.BaseURL)
	}
	if cfg.DatabasePath != "/data/app.db" {
		t.Errorf("DatabasePath = %q", cfg.DatabasePath)
	}
	if cfg.SessionSecret != "test-secret" {
		t.Errorf("SessionSecret = %q", cfg.SessionSecret)
	}
	if cfg.ResendAPIKey != "re_xxx" {
		t.Errorf("ResendAPIKey = %q", cfg.ResendAPIKey)
	}
	if cfg.EmailFrom != "noreply@example.com" {
		t.Errorf("EmailFrom = %q", cfg.EmailFrom)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q", cfg.LogLevel)
	}
}

func TestIsDevelopment(t *testing.T) {
	cfg := &Config{Env: "development"}
	if !cfg.IsDevelopment() {
		t.Error("expected IsDevelopment() == true for Env=development")
	}
	cfg.Env = "production"
	if cfg.IsDevelopment() {
		t.Error("expected IsDevelopment() == false for Env=production")
	}
}
```

- [ ] **Шаг 2: Запустить тест и убедиться что падает**

```bash
go test ./internal/config/...
```

Ожидается: ошибка компиляции (`undefined: Load`, `undefined: Config`).

- [ ] **Шаг 3: Написать реализацию**

Создать `internal/config/config.go`:

```go
package config

import "os"

type Config struct {
	Env           string
	Port          string
	BaseURL       string
	DatabasePath  string
	SessionSecret string
	ResendAPIKey  string
	EmailFrom     string
	LogLevel      string
}

func Load() (*Config, error) {
	cfg := &Config{
		Env:           getEnv("ENV", "development"),
		Port:          getEnv("PORT", "8080"),
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),
		DatabasePath:  getEnv("DATABASE_PATH", "./app.db"),
		SessionSecret: getEnv("SESSION_SECRET", "dev-secret-not-for-production"),
		ResendAPIKey:  os.Getenv("RESEND_API_KEY"),
		EmailFrom:     getEnv("EMAIL_FROM", "noreply@localhost"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
	}
	return cfg, nil
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Шаг 4: Запустить тест**

```bash
go test ./internal/config/... -v
```

Ожидается: `PASS`, 3 теста.

- [ ] **Шаг 5: Закоммитить**

```bash
git add internal/config/
git commit -m "Добавить пакет config с загрузкой из env-переменных"
```

---

## Task 3: Middleware `recover` (TDD)

**Files:**
- Create: `internal/http/middleware/recover.go`
- Create: `internal/http/middleware/recover_test.go`

Middleware ловит panic и возвращает 500 вместо падения сервера.

- [ ] **Шаг 1: Написать тест**

Создать `internal/http/middleware/recover_test.go`:

```go
package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecover_PanicReturns500(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := Recover(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

func TestRecover_NoPanicPassesThrough(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := Recover(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want 418", rec.Code)
	}
}
```

- [ ] **Шаг 2: Запустить тест (должен упасть компиляцией)**

```bash
go test ./internal/http/middleware/...
```

- [ ] **Шаг 3: Написать реализацию**

Создать `internal/http/middleware/recover.go`:

```go
package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered",
						"error", rec,
						"path", r.URL.Path,
						"stack", string(debug.Stack()),
					)
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
```

- [ ] **Шаг 4: Запустить тест**

```bash
go test ./internal/http/middleware/... -v
```

Ожидается: 2 теста PASS.

- [ ] **Шаг 5: Закоммитить**

```bash
git add internal/http/middleware/
git commit -m "Добавить middleware Recover для отлова panic"
```

---

## Task 4: Health-check handler (TDD)

**Files:**
- Create: `internal/http/handlers/health.go`
- Create: `internal/http/handlers/health_test.go`

- [ ] **Шаг 1: Написать тест**

Создать `internal/http/handlers/health_test.go`:

```go
package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if body := rec.Body.String(); body != "ok" {
		t.Errorf("body = %q, want %q", body, "ok")
	}
}
```

- [ ] **Шаг 2: Запустить тест**

```bash
go test ./internal/http/handlers/...
```

Ожидается: компиляция не проходит.

- [ ] **Шаг 3: Написать реализацию**

Создать `internal/http/handlers/health.go`:

```go
package handlers

import "net/http"

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
```

- [ ] **Шаг 4: Запустить тест**

```bash
go test ./internal/http/handlers/... -v
```

Ожидается: PASS.

- [ ] **Шаг 5: Закоммитить**

```bash
git add internal/http/handlers/
git commit -m "Добавить health-check handler"
```

---

## Task 5: Точка входа `cmd/server/main.go`

**Files:**
- Create: `cmd/server/main.go`
- Modify: `go.mod` (через `go get`)

Связывает конфиг, роутер, middleware и запускает HTTP-сервер с graceful shutdown.

- [ ] **Шаг 1: Добавить зависимости chi и godotenv**

```bash
go get github.com/go-chi/chi/v5@latest
go get github.com/joho/godotenv@latest
```

- [ ] **Шаг 2: Написать `cmd/server/main.go`**

Создать `cmd/server/main.go`:

```go
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"secret-santa/internal/config"
	"secret-santa/internal/http/handlers"
	"secret-santa/internal/http/middleware"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	logger := newLogger(cfg)
	slog.SetDefault(logger)

	r := chi.NewRouter()
	r.Use(chimid.RequestID)
	r.Use(middleware.Recover(logger))
	r.Use(chimid.StripSlashes)

	r.Get("/healthz", handlers.Health)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("server listening", "addr", srv.Addr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen error", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "err", err)
	}
}

func newLogger(cfg *config.Config) *slog.Logger {
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
	opts := &slog.HandlerOptions{Level: level}
	if cfg.IsDevelopment() {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
```

- [ ] **Шаг 3: Запустить сервер и проверить health-check**

В одном терминале:
```bash
go run ./cmd/server
```

Ожидаемый вывод: `level=INFO msg="server listening" addr=:8080 env=development`

В другом терминале:
```bash
curl -i http://localhost:8080/healthz
```

Ожидается: `HTTP/1.1 200 OK` и тело `ok`.

Остановить сервер `Ctrl+C`, должен увидеть `shutting down`.

- [ ] **Шаг 4: Прогнать все тесты**

```bash
go test ./...
```

Ожидается: PASS во всех пакетах.

- [ ] **Шаг 5: Закоммитить**

```bash
git add cmd/ go.mod go.sum
git commit -m "Добавить main.go с chi-роутером и graceful shutdown"
```

---

## Task 6: Инициализация фронта (Vite + React + TS)

**Files:**
- Create: `web/package.json`
- Create: `web/tsconfig.json`
- Create: `web/tsconfig.node.json`
- Create: `web/vite.config.ts`
- Create: `web/index.html`
- Create: `web/src/main.tsx`
- Create: `web/src/App.tsx`
- Create: `web/src/vite-env.d.ts`

- [ ] **Шаг 1: Создать скаффолд через Vite**

```bash
cd web
npm create vite@latest . -- --template react-ts
```

На вопрос «Current directory is not empty» выбрать `Ignore files and continue`.

- [ ] **Шаг 2: Установить зависимости**

```bash
npm install
```

- [ ] **Шаг 3: Установить react-router**

```bash
npm install react-router@7
```

- [ ] **Шаг 4: Проверить, что дев-сервер поднимается**

```bash
npm run dev
```

Ожидается: `VITE v5.x.x ready in ...ms`, слушает на `http://localhost:5173`.

Открыть в браузере — должна быть стандартная страница Vite+React. Остановить `Ctrl+C`.

- [ ] **Шаг 5: Заменить `web/src/App.tsx` на минимальный каркас**

Переписать `web/src/App.tsx`:

```tsx
import { BrowserRouter, Routes, Route } from 'react-router'
import './App.css'

function Landing() {
  return (
    <main className="min-h-screen flex items-center justify-center">
      <h1 className="text-3xl font-bold">Тайный Санта</h1>
    </main>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Landing />} />
      </Routes>
    </BrowserRouter>
  )
}
```

- [ ] **Шаг 6: Очистить `web/src/App.css` и `web/src/index.css`**

Перезаписать `web/src/App.css` пустым содержимым:
```css
```

Перезаписать `web/src/index.css`:
```css
:root {
  font-family: system-ui, -apple-system, sans-serif;
  line-height: 1.5;
}
* { box-sizing: border-box; }
body { margin: 0; }
```

- [ ] **Шаг 7: Закоммитить**

```bash
cd ..
git add web/
git commit -m "Инициализировать фронт (Vite + React + TS + react-router)"
```

---

## Task 7: Tailwind CSS

**Files:**
- Create: `web/tailwind.config.js`
- Create: `web/postcss.config.js`
- Modify: `web/src/index.css`
- Modify: `web/package.json` (через `npm install`)

- [ ] **Шаг 1: Установить Tailwind v3**

```bash
cd web
npm install -D tailwindcss@3 postcss autoprefixer
npx tailwindcss init -p
```

Это создаст `tailwind.config.js` и `postcss.config.js`.

- [ ] **Шаг 2: Настроить `tailwind.config.js`**

Переписать `web/tailwind.config.js`:

```js
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
```

- [ ] **Шаг 3: Подключить Tailwind в `index.css`**

Переписать `web/src/index.css`:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

- [ ] **Шаг 4: Проверить работу Tailwind**

```bash
npm run dev
```

Открыть `http://localhost:5173`. Заголовок «Тайный Санта» должен быть крупным и жирным (классы `text-3xl font-bold` уже в App.tsx).

Остановить `Ctrl+C`.

- [ ] **Шаг 5: Закоммитить**

```bash
cd ..
git add web/tailwind.config.js web/postcss.config.js web/src/index.css web/package.json web/package-lock.json
git commit -m "Подключить Tailwind CSS"
```

---

## Task 8: Vite-прокси для `/api` и `/ws` на Go-сервер

**Files:**
- Modify: `web/vite.config.ts`

- [ ] **Шаг 1: Переписать `web/vite.config.ts`**

```ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: false,
      },
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
        changeOrigin: false,
      },
      '/healthz': {
        target: 'http://localhost:8080',
        changeOrigin: false,
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
```

- [ ] **Шаг 2: Проверить прокси**

В одном терминале:
```bash
go run ./cmd/server
```

В другом:
```bash
cd web
npm run dev
```

В третьем:
```bash
curl -i http://localhost:5173/healthz
```

Ожидается: `HTTP/1.1 200 OK`, тело `ok` (прокси передал запрос на Go).

Остановить оба процесса.

- [ ] **Шаг 3: Закоммитить**

```bash
cd ..
git add web/vite.config.ts
git commit -m "Настроить Vite-прокси для /api, /ws и /healthz"
```

---

## Task 9: Embed статики фронта в Go-бинарник

**Files:**
- Create: `web/embed.go`
- Create: `web/dist/.gitkeep`
- Modify: `.gitignore`
- Modify: `cmd/server/main.go`

Go-сервер должен отдавать `web/dist/` в проде. В dev-режиме директория пустая — статика отдается Vite на порту 5173. `//go:embed` ищет файлы относительно пакета, где объявлена директива, поэтому embed-файл лежит в `web/` (пакет `web`), импортируется из `cmd/server/main.go`.

- [ ] **Шаг 1: Создать заглушку `web/dist/.gitkeep`**

```bash
mkdir -p web/dist
touch web/dist/.gitkeep
```

Это нужно, чтобы `//go:embed all:dist` не падал на пустой папке.

- [ ] **Шаг 2: Поправить `.gitignore`, чтобы `.gitkeep` попал в коммит**

Переписать `.gitignore`, заменив строку `web/dist/` на:

```
web/dist/*
!web/dist/.gitkeep
```

Полностью `.gitignore` теперь выглядит так:

```
# Go
*.exe
*.test
*.out
vendor/

# Node / Vite
node_modules/
web/dist/*
!web/dist/.gitkeep
.DS_Store

# Env
.env
.env.local

# Local DB
*.db
*.db-journal

# Editor
.vscode/
.idea/
*.swp
```

- [ ] **Шаг 3: Создать `web/embed.go`**

```go
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

func FS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
```

- [ ] **Шаг 4: Проверить, что пакет компилируется**

```bash
go build ./web/...
```

Ожидается: команда завершается без ошибок (TypeScript-файлы в `web/src/` Go-компилятор игнорирует — они не `.go`).

- [ ] **Шаг 5: Подключить frontendFS в main.go**

Обновить `cmd/server/main.go` — добавить импорт и отдачу статики. Заменить секцию, где регистрируется маршрут `/healthz` и добавить fallback на статику **только если frontendFS содержит `index.html`** (в dev он пустой, тогда пропускаем).

Полностью переписать `cmd/server/main.go`:

```go
package main

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"secret-santa/internal/config"
	"secret-santa/internal/http/handlers"
	"secret-santa/internal/http/middleware"
	"secret-santa/web"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	logger := newLogger(cfg)
	slog.SetDefault(logger)

	r := chi.NewRouter()
	r.Use(chimid.RequestID)
	r.Use(middleware.Recover(logger))
	r.Use(chimid.StripSlashes)

	r.Get("/healthz", handlers.Health)

	if err := mountFrontend(r, logger); err != nil {
		logger.Warn("frontend not mounted (dev mode?)", "err", err)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("server listening", "addr", srv.Addr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen error", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "err", err)
	}
}

func mountFrontend(r chi.Router, logger *slog.Logger) error {
	frontendFS, err := web.FS()
	if err != nil {
		return err
	}
	if _, err := fs.Stat(frontendFS, "index.html"); err != nil {
		return err
	}

	fileServer := http.FileServer(http.FS(frontendFS))
	r.Handle("/assets/*", fileServer)
	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		// SPA fallback: все, что не /api и не /healthz — отдаем index.html
		req.URL.Path = "/"
		fileServer.ServeHTTP(w, req)
	})
	logger.Info("frontend mounted from embedded FS")
	return nil
}

func newLogger(cfg *config.Config) *slog.Logger {
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
	opts := &slog.HandlerOptions{Level: level}
	if cfg.IsDevelopment() {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
```

- [ ] **Шаг 6: Проверить сборку**

```bash
go build ./...
```

Ожидается: компиляция без ошибок.

- [ ] **Шаг 7: Запустить — убедиться, что в dev-режиме сервер стартует без фронта**

```bash
go run ./cmd/server
```

Ожидается лог: `frontend not mounted (dev mode?)` и `server listening`. `curl http://localhost:8080/healthz` → `ok`.

Остановить.

- [ ] **Шаг 8: Закоммитить**

```bash
git add .gitignore web/embed.go web/dist/.gitkeep cmd/server/main.go
git commit -m "Подключить embed статики фронта с dev-fallback"
```

---

## Task 10: Полный цикл dev-сборки — `make dev` и `make build`

**Files:**
- Create: `Makefile`

- [ ] **Шаг 1: Создать `Makefile`**

```makefile
.PHONY: dev dev-server dev-web build test fmt clean

# Запускает бэкенд и фронт параллельно
dev:
	@echo "Запуск бэка на :8080 и фронта на :5173"
	@trap 'kill 0' INT TERM; \
		(go run ./cmd/server) & \
		(cd web && npm run dev) & \
		wait

dev-server:
	go run ./cmd/server

dev-web:
	cd web && npm run dev

# Прод-сборка: собирает фронт в web/dist, затем Go-бинарник
build:
	cd web && npm ci && npm run build
	CGO_ENABLED=0 go build -o bin/server ./cmd/server

test:
	go test ./... -v
	cd web && npm run build

fmt:
	go fmt ./...

clean:
	rm -rf bin web/dist/assets web/dist/index.html
```

- [ ] **Шаг 2: Проверить `make test`**

```bash
make test
```

Ожидается: все Go-тесты PASS, `npm run build` успешно создает `web/dist/index.html` и `web/dist/assets/...`.

- [ ] **Шаг 3: Проверить прод-сборку локально**

```bash
make build
./bin/server
```

Ожидается: `frontend mounted from embedded FS` в логах. Открыть `http://localhost:8080/` — должен быть «Тайный Санта» (React через embed).

Остановить сервер `Ctrl+C`.

- [ ] **Шаг 4: Очистить собранный фронт перед коммитом**

```bash
make clean
```

- [ ] **Шаг 5: Закоммитить Makefile**

```bash
git add Makefile
git commit -m "Добавить Makefile: make dev, make build, make test"
```

---

## Task 11: Multi-stage Dockerfile

**Files:**
- Create: `Dockerfile`
- Create: `.dockerignore`

- [ ] **Шаг 1: Создать `.dockerignore`**

```
.git
.github
.vscode
.idea
bin/
web/node_modules/
web/dist/
*.db
*.db-journal
.env
.env.local
docs/
```

- [ ] **Шаг 2: Создать `Dockerfile`**

```dockerfile
# syntax=docker/dockerfile:1.6

# Stage 1: build frontend
FROM node:20-alpine AS web
WORKDIR /build
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: build Go binary with embedded frontend
FROM golang:1.22-alpine AS server
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN rm -rf web/dist
COPY --from=web /build/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

# Stage 3: minimal runtime
FROM alpine:3.20
RUN adduser -D -u 1000 app && apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=server /out/server /usr/local/bin/server
USER app
EXPOSE 8080
ENV PORT=8080 ENV=production
CMD ["/usr/local/bin/server"]
```

- [ ] **Шаг 3: Собрать образ**

```bash
docker build -t secret-santa:dev .
```

Ожидается: успешная сборка (занимает 1-3 минуты в первый раз).

- [ ] **Шаг 4: Запустить контейнер**

```bash
docker run --rm -p 8080:8080 -e ENV=production secret-santa:dev
```

В другом терминале:
```bash
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/
```

Ожидается:
- `/healthz` → 200 `ok`
- `/` → 200, HTML с заголовком «Тайный Санта»

Остановить контейнер `Ctrl+C`.

- [ ] **Шаг 5: Закоммитить**

```bash
git add Dockerfile .dockerignore
git commit -m "Добавить multi-stage Dockerfile"
```

---

## Task 12: README с инструкциями для разработки

**Files:**
- Modify: `README.md`

- [ ] **Шаг 1: Переписать `README.md`**

```markdown
# Тайный Санта

Веб-приложение для организации игры «Тайный Санта» среди друзей: создание группы, регистрация участников, случайное распределение, анонимный чат в реальном времени.

Дизайн проекта: [`docs/project-design.md`](docs/project-design.md).

## Требования

- Go 1.22+
- Node.js 20+
- (опционально) Docker для прод-сборки

## Разработка

```bash
# Первый раз — установить зависимости фронта
cd web && npm install && cd ..

# Запустить бэк (:8080) и фронт (:5173) параллельно
make dev
```

Открыть `http://localhost:5173` — Vite проксирует `/api/*`, `/ws/*`, `/healthz` на бэкенд.

## Переменные окружения

Скопировать `.env.example` в `.env` и при необходимости поправить значения. В dev все переменные имеют безопасные значения по умолчанию, `.env` не обязателен.

## Сборка

```bash
make build      # собирает фронт + Go-бинарник в bin/server
make test       # прогоняет все тесты
docker build -t secret-santa .
```

## Структура проекта

```
cmd/server/          Точка входа бэка
internal/            Доменные пакеты (config, http, auth, groups, draw, chat, ...)
web/                 Фронт (React + Vite + TS + Tailwind)
web/dist/            Артефакты сборки фронта (встраиваются в Go-бинарник)
docs/                Дизайн-документы и планы
```
```

- [ ] **Шаг 2: Создать `.env.example`**

```
# Окружение: development | production
ENV=development

# Порт HTTP-сервера
PORT=8080

# Базовый URL для формирования magic-link
BASE_URL=http://localhost:8080

# Путь к SQLite-файлу (используется в следующих планах)
DATABASE_PATH=./app.db

# Секрет для подписи session-токенов (32 случайных байта в base64)
SESSION_SECRET=change-me-in-production

# Ключ Resend API для отправки писем (пустое значение в dev — письма печатаются в лог)
RESEND_API_KEY=

# Адрес отправителя
EMAIL_FROM=noreply@localhost

# Уровень логирования: debug | info | warn | error
LOG_LEVEL=info
```

- [ ] **Шаг 3: Закоммитить**

```bash
git add README.md .env.example
git commit -m "Обновить README и добавить .env.example"
```

---

## Task 13: CI — GitHub Actions (smoke)

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Шаг 1: Создать workflow**

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: web/package-lock.json

      - name: Install frontend deps
        working-directory: web
        run: npm ci

      - name: Build frontend (для embed)
        working-directory: web
        run: npm run build

      - name: Go test
        run: go test ./... -race

      - name: Go build
        run: go build ./...

      - name: Typecheck frontend
        working-directory: web
        run: npm run build
```

- [ ] **Шаг 2: Закоммитить**

```bash
git add .github/
git commit -m "Добавить GitHub Actions CI"
```

---

## Финальная проверка плана

- [ ] Все задачи выполнены, все коммиты сделаны.
- [ ] `make test` проходит локально.
- [ ] `make dev` запускает фронт + бэк, `http://localhost:5173` открывает страницу «Тайный Санта».
- [ ] `make build && ./bin/server` — запускает прод-режим, `http://localhost:8080` открывает ту же страницу (без Vite).
- [ ] `docker build` и `docker run` работают.

## Что дальше

**План 2 — DB + Auth**: подключить SQLite, `sqlc`, `golang-migrate`, написать миграции для `users/sessions/magic_links`, реализовать magic-link flow на бэке и фронте.
