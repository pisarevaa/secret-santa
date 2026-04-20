# Тайный Санта

Веб-приложение для организации игры «Тайный Санта». Организатор создает группу,
делится ссылкой-приглашением, участники регистрируются и указывают вишлист.
Организатор запускает жеребьевку, после чего каждый видит своего подопечного
и может анонимно переписываться с ним и со своим Сантой в реальном времени.

MVP для личного использования: до 10 активных групп, до 20 человек в группе,
один инстанс, SQLite в файле. UI только на русском.

## Стек

**Бэкенд (Go 1.26):**
- `chi` — HTTP-роутер
- `coder/websocket` — WebSocket для чата
- `modernc.org/sqlite` (pure-Go, без CGO) + `sqlc` для типобезопасных запросов
- `golang-migrate` — миграции, применяются на старте
- `slog` — логи
- Resend API для email с magic-link (в dev-режиме — вывод в лог)

**Фронтенд (TypeScript):**
- React 19 + Vite
- `react-router` v7
- Tailwind CSS
- Нативные `fetch` и `WebSocket`

**Деплой:** Fly.io, multi-stage Dockerfile, persistent volume для SQLite.

## Структура

```
secret-santa/
├── cmd/server/main.go       # точка входа
├── internal/
│   ├── auth/                # magic link, сессии
│   ├── chat/                # WebSocket-хаб, сообщения
│   ├── config/              # env-переменные
│   ├── db/                  # sqlite, миграции, sqlc
│   ├── draw/                # алгоритм жеребьевки
│   ├── email/               # Resend + dev-заглушка
│   ├── groups/              # создание, join, membership
│   └── http/                # handlers, middleware
├── web/                     # React + Vite
├── docs/                    # дизайн и план реализации
├── embed.go                 # go:embed web/dist
├── Dockerfile
└── fly.toml
```

Монолит: Go-сервер обслуживает REST API, WebSocket и статику фронта с одного
домена. В dev-режиме Vite на `:5173` проксирует `/api/*` и `/ws/*` на Go
на `:8080`. В проде `web/dist` встраивается в бинарник через `//go:embed`.

## Запуск локально

Требуется Go 1.26+ и Node.js 20+.

1. Скопировать `.env.example` в `.env` и при необходимости поправить значения:
   ```sh
   cp .env.example .env
   ```

2. Поставить зависимости фронта:
   ```sh
   cd web && npm install && cd ..
   ```

3. Запустить бэкенд (в одном терминале):
   ```sh
   go run ./cmd/server
   ```
   Сервер поднимется на `:8080`, применит миграции и создаст `app.db`.

4. Запустить фронт в dev-режиме (во втором терминале):
   ```sh
   cd web && npm run dev
   ```
   Приложение будет доступно на `http://localhost:5173`.

В dev-окружении magic-link не отправляется на почту, а печатается в лог
сервера — ссылку можно скопировать оттуда.

## Переменные окружения

| Переменная       | Назначение                                      | Значение по умолчанию      |
|------------------|--------------------------------------------------|----------------------------|
| `PORT`           | Порт HTTP-сервера                                | `8080`                     |
| `BASE_URL`       | Публичный URL для magic-link                     | `http://localhost:5173`    |
| `DATABASE_PATH`  | Путь к файлу SQLite                              | `app.db`                   |
| `SESSION_SECRET` | Секрет для подписи сессионных куки               | `dev-secret-change-me`     |
| `RESEND_API_KEY` | Ключ Resend API (пусто в dev)                    | —                          |
| `EMAIL_FROM`     | Адрес отправителя писем                          | `noreply@localhost`        |
| `ENV`            | `development` или `production`                   | `development`              |
| `LOG_LEVEL`      | `debug` / `info` / `warn` / `error`              | `debug`                    |

## Тесты

```sh
go test ./...             # юнит- и интеграционные тесты на Go
cd web && npm run lint    # ESLint фронта
cd web && npm run build   # typecheck + сборка
```

В CI (`.github/workflows/ci.yml`) прогоняются те же проверки.

## Сборка и деплой

Прод-сборка одной командой через Docker:

```sh
docker build -t secret-santa .
docker run --rm -p 8080:8080 -v $(pwd)/data:/data \
  -e DATABASE_PATH=/data/app.db secret-santa
```

Multi-stage: Node собирает фронт → Go компилирует бинарник с встроенной
статикой → финальный alpine-образ.

Деплой на Fly.io:

```sh
fly deploy                                        # конфиг в fly.toml
fly secrets set SESSION_SECRET=... RESEND_API_KEY=... EMAIL_FROM=...
```

## Документация

- [`docs/project-design.md`](docs/project-design.md) — дизайн и архитектура.
- [`docs/project-implementation.md`](docs/project-implementation.md) — план
  реализации по фичам и отчеты о выполнении.
