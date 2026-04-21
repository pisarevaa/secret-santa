# Тайный Санта — дизайн проекта

**Дата:** 2026-04-16
**Статус:** утвержден к реализации

## 1. Цель и рамки

Веб-приложение для организации игры «Тайный Санта» среди друзей. Организатор создает группу, делится общей ссылкой-приглашением, участники регистрируются и указывают вишлист. Организатор вручную запускает жеребьевку. После жеребьевки каждый видит своего подопечного и может анонимно переписываться с ним и со своим Сантой в реальном времени.

### Масштаб и критерии успеха

- MVP для личного использования, до 10 активных групп одновременно, до 20 человек в группе.
- Работает и не стыдно показать друзьям.
- Один регион, один инстанс, SQLite в файле.

### Явно вне объема (YAGNI)

- Аналитика, метрики, Sentry.
- Email-уведомления о новых сообщениях (только in-app).
- Мобильное приложение / PWA.
- Шардирование, несколько регионов.
- Редактирование и удаление сообщений, файлы, реакции.
- i18n — UI только на русском.
- Правила жеребьевки (исключения пар, подгруппы).
- Автоматический cleanup старых групп.

## 2. Стек

**Бэкенд (Go 1.26):**
- Роутер: `chi`
- WebSocket: `coder/websocket`
- БД: SQLite через `modernc.org/sqlite` (pure-Go, без CGO) + `sqlc` для запросов
- Миграции: `golang-migrate`, применяются на старте
- Email: прямой HTTP-запрос к Resend API
- Логи: `slog` (stdlib)
- Конфиг: переменные окружения

**Фронтенд (TypeScript):**
- React 18 + Vite
- Роутинг: `react-router` v7
- Стили: Tailwind CSS
- HTTP: нативный `fetch` + тонкая обертка
- WebSocket: нативный `WebSocket` API с авто-реконнектом
- Состояние: `useState` / `useReducer`, без Redux

**Деплой:**
- Fly.io, один `shared-cpu-1x` инстанс, один регион.
- Multi-stage Dockerfile: Node собирает фронт → Go компилирует бинарник с фронтом внутри (`//go:embed`) → финальный alpine-образ.
- Persistent volume 1 GB для SQLite.
- Секреты через `fly secrets set`.

## 3. Архитектура и структура

Монолит: Go-сервер обслуживает REST API, WebSocket и статические файлы фронта с одного домена. CORS не нужен.

```
secret-santa/
├── cmd/server/main.go
├── internal/
│   ├── config/              # env-переменные
│   ├── db/                  # sqlite-подключение, миграции
│   │   ├── migrations/*.sql
│   │   └── queries/*.sql    # sqlc
│   ├── auth/                # magic link, сессии
│   ├── groups/              # создание, join, membership
│   ├── draw/                # чистый алгоритм жеребьевки
│   ├── chat/                # WebSocket-хаб, сообщения
│   ├── email/               # Resend-клиент + dev-заглушка
│   └── http/
│       ├── handlers/
│       └── middleware/
├── web/
│   ├── src/
│   │   ├── pages/
│   │   ├── components/
│   │   ├── api/
│   │   └── main.tsx
│   └── vite.config.ts
├── embed.go                 # go:embed web/dist
├── Dockerfile
├── fly.toml
└── README.md
```

**Принцип границ:** пакеты в `internal/` содержат чистую доменную логику, не знают про HTTP. HTTP-handlers парсят запросы, вызывают доменные функции, форматируют ответы. Это делает домен тестируемым без поднятия сервера.

**Dev-поток:** Vite на `:5173` проксирует `/api/*` и `/ws/*` на Go-сервер `:8080`. В проде фронт собирается в `web/dist` и встраивается в бинарник.

## 4. Модель данных (SQLite)

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
    status       TEXT NOT NULL,           -- 'open' | 'drawn'
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
    token       TEXT PRIMARY KEY,          -- 32 байта base64url
    email       TEXT NOT NULL,
    expires_at  DATETIME NOT NULL,         -- 15 минут
    used_at     DATETIME
);

CREATE TABLE sessions (
    token       TEXT PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id),
    expires_at  DATETIME NOT NULL,         -- 30 дней
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE messages (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id       INTEGER NOT NULL REFERENCES groups(id),
    sender_id      INTEGER NOT NULL REFERENCES users(id),
    recipient_id   INTEGER NOT NULL REFERENCES users(id),
    direction      TEXT NOT NULL,          -- 'santa_to_recipient' | 'recipient_to_santa'
    body           TEXT NOT NULL,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_pair ON messages(group_id, sender_id, recipient_id, created_at);
CREATE INDEX idx_memberships_group ON memberships(group_id);
```

**Заметки:**
- `direction` различает два чата одного пользователя: один как Санта, другой как подопечный. Тройка `(group_id, sender_id, recipient_id)` однозначно идентифицирует диалог.
- Имена в чате скрыты на уровне отображения: в БД реальные `user_id`, фронт рисует «Твой Санта» / «Твой подопечный». Шифровать на уровне БД не нужно.

## 5. API

### REST

```
POST   /api/auth/request-link     {email}                → 204 (всегда, не утечка наличия email)
GET    /api/auth/verify?token=... → ставит cookie, 302 на /
POST   /api/auth/logout           → удаляет сессию
GET    /api/auth/me               → {user_id, email, name} или 401

POST   /api/groups                {title}                → {id, invite_code}           (сессия)
GET    /api/groups/:invite_code   → если сессии нет: {title, member_count, status};
                                     если есть и юзер — участник: {title, status, members:[{name, is_me}], is_organizer, my_membership_id}
POST   /api/groups/:invite_code/join  {name, wishlist}   → 204                         (сессия)
PATCH  /api/memberships/:id       {wishlist}             → 204                         (свой membership)
POST   /api/groups/:id/draw       → 204                  (организатор, status=open)
GET    /api/groups/:id/my-recipient  → {recipient: {name, wishlist}}  (после draw)

GET    /api/groups/:id/chats/:role?before=...  → [{id, from_me, body, created_at}...]
         :role = 'santa' (мой чат как Санта с подопечным) | 'recipient' (мой чат как подопечный с Сантой)
```

### WebSocket

```
WS /ws/groups/:id    (авторизация по session-cookie при handshake)

Client → Server:
  {"type":"send", "role":"santa", "body":"..."}
     # role = 'santa'     — пишу своему подопечному (я — Санта)
     # role = 'recipient' — пишу своему Санте        (я — подопечный)

Server → Client:
  {"type":"message", "id":..., "role":"santa"|"recipient", "from_me":bool, "body":"...", "created_at":"..."}
     # role здесь — в каком из двух моих чатов это сообщение (с точки зрения получателя)
  {"type":"drawn"}
  {"type":"error", "reason":"..."}
```

Один WS на группу. Клиент получает сообщения обоих своих чатов (как Санта и как подопечный), поле `role` позволяет фронту положить сообщение в нужный чат.

**Маппинг `role` ↔ `messages.direction`:** role и direction — это одно и то же понятие с разных точек зрения. Для сообщения, отправленного пользователем U с `role='santa'`: в БД пишется `direction='santa_to_recipient'`, `sender_id=U`, `recipient_id=<подопечный U>`. Для `role='recipient'`: `direction='recipient_to_santa'`, `recipient_id=<Санта U>`. При чтении сервер вычисляет `role` для каждого сообщения исходя из позиции зрителя.

### Формат ошибок

```
HTTP: 400 | 401 | 403 | 404 | 409 | 429 | 500
Body: {"error": "code", "message": "человеко-читаемое"}
```

Коды: `invalid_input`, `unauthorized`, `forbidden`, `not_found`, `already_drawn`, `not_enough_members`, `rate_limited`.

### Лимиты на входе

- `wishlist` ≤ 2000 символов
- сообщение ≤ 2000 символов
- `title` группы ≤ 100 символов
- `name` участника ≤ 50 символов
- email — валидация формата

## 6. Пользовательские сценарии

1. **Создание группы.** Организатор заходит на `/` → вводит email → получает magic link → переходит по нему → создает группу → получает ссылку `app.fly.dev/g/ABC123`.
2. **Регистрация участника.** Переходит по ссылке → вводит email → magic link → возвращается → вводит имя и вишлист → видит список участников (без пар).
3. **Жеребьевка.** Организатор нажимает «Провести». В одной транзакции распределяются пары и группа переходит в `status='drawn'`. Все открытые WS получают `{"type":"drawn"}`, фронты обновляют страницу.
4. **Просмотр и чат.** Страница группы после жеребьевки показывает «Твой подопечный» (имя + вишлист) и два чата: «Переписка с твоим Сантой» (анонимно) и «Переписка с твоим подопечным» (анонимно).

## 7. Жеребьевка

Чистая функция в `internal/draw`:

```go
func Assign(participants []int64, rng *rand.Rand) map[int64]int64
```

**Алгоритм — рандомизированный цикл:**
1. Если участников меньше 2 → ошибка `not_enough_members`.
2. Перемешать массив через `rng.Shuffle` (Fisher–Yates).
3. `perm[i] → perm[(i+1) % n]` — каждый дарит следующему по кругу.

**Свойства:** никто не дарит себе, у каждого ровно один подопечный и ровно один Санта, один большой цикл (в малых группах это уменьшает возможности угадать).

**Handler `POST /api/groups/:id/draw`:**
- Читает `user_id` из всех `memberships` группы.
- Вызывает `Assign`.
- В одной транзакции: обновляет `recipient_id` у каждого membership + `groups.status='drawn'`, `drawn_at=NOW()`.
- Условие `WHERE status='open'` в `UPDATE groups` защищает от гонок: если уже `drawn` — 409.
- После коммита — бродкаст `{"type":"drawn"}` в WS-хаб группы.

## 8. Чат — архитектура

**Hub-per-group** в памяти — подходит для одного инстанса.

```go
type Hub struct {
    groupID     int64
    clients     map[int64]map[*Client]struct{}  // userID → активные соединения
    register    chan *Client
    unregister  chan *Client
    incoming    chan inboundMessage
}
```

**Поток сообщения:**
1. Клиент открывает `WS /ws/groups/:id`. Middleware проверяет session-cookie, извлекает `userID`, проверяет что есть membership в группе и `status='drawn'`. Иначе close-код 1008.
2. Handler регистрирует `Client` в `Hub` группы (создает, если нет).
3. Клиент шлет `{"type":"send", "role":"...", "body":"..."}`.
4. Hub по `role` и `sender_id` определяет `recipient_id` из кэшированных memberships (для `role=santa` это подопечный отправителя, для `role=recipient` — его Санта), сохраняет сообщение в `messages` с соответствующим `direction`, рассылает обоим участникам диалога (отправителю тоже — как подтверждение). Перед отправкой сервер вычисляет `role` с точки зрения каждого получателя.
5. Клиенты других пар в группе сообщение не получают — Hub фильтрует по `(sender_id, recipient_id)`.

**История:** REST `GET /api/groups/:id/chats/:direction` отдает последние 50 сообщений. Дальше WS догружает новое.

**Лимиты:** сообщение ≤ 2000 символов; rate 10 сообщений/мин на пользователя (в Hub, простой токен-бакет).

**Graceful shutdown:** на SIGTERM — close всех WS с кодом 1001, ожидание до 10 сек.

## 9. Аутентификация

**Magic link:**
- `POST /api/auth/request-link` с email.
- Генерируется токен (32 байта `crypto/rand`, base64url), сохраняется в `magic_links` с `expires_at = now+15m`.
- Письмо через Resend API: `{BASE_URL}/api/auth/verify?token=...`.
- В dev-режиме (`ENV=development`) письма печатаются в лог вместо отправки.

**Verify:**
- `GET /api/auth/verify?token=...`.
- Проверяем: токен существует, не истек, `used_at IS NULL`.
- Проставляем `used_at=NOW()`. Если `users` с таким email нет — создаем.
- Генерируем session-токен, сохраняем в `sessions`, ставим cookie `s=...; HttpOnly; Secure; SameSite=Lax; Max-Age=30d`.
- 302 на `/`.

**Middleware `RequireSession`:** читает cookie `s`, по токену ищет `sessions`, проверяет `expires_at`, кладет `userID` в контекст.

## 10. Обработка ошибок и надежность

- **Валидация на входе** — внутри каждого handler перед бизнес-логикой.
- **Panic recovery** в middleware — один упавший handler не убивает сервер.
- **Rate limits** (`golang.org/x/time/rate`, bucket в памяти):
  - `/auth/request-link`: 5 запросов на email+IP за 10 минут.
  - Создание группы: 3 в час на пользователя.
  - Сообщения в чат: 10 в минуту на пользователя (в Hub).
- **WebSocket reconnect:** клиент делает экспоненциальный backoff до 30 сек с jitter.
- **Миграции** применяются на старте сервера. При ошибке — сервер не стартует.
- **Health-check:** `GET /healthz` возвращает 200 без проверки БД (Fly.io смотрит только что порт отвечает).

## 11. Тестирование

- **`internal/draw`** — юниты: детерминированный `rand` с фиксированным seed, проверка инвариантов (никто себе, все назначены, один цикл).
- **`internal/auth`** — юниты: выпуск и валидация magic link, expiry, повторное использование, сессии.
- **`internal/groups`** — интеграционные тесты на `:memory:` SQLite: create, join, draw, ограничения.
- **`internal/chat`** — тесты Hub с mock-клиентами: регистрация, доставка по направлению, фильтрация, rate limit.
- **HTTP-handlers** — `httptest.Server` + in-memory DB, ~5 smoke-тестов на главные сценарии (регистрация → создание группы → join → draw → отправка сообщения).
- **Фронт** — ручное тестирование в MVP. E2E (Playwright) — отдельная задача, если будет нужно.
- **CI (GitHub Actions):** `go test ./...` + `npm run typecheck` + `npm run build` (фронт должен собираться).

## 12. Конфигурация (env)

```
BASE_URL           https://secret-santa.fly.dev
DATABASE_PATH      /data/app.db
SESSION_SECRET     <random-32-bytes-base64>
RESEND_API_KEY     re_xxx
EMAIL_FROM         noreply@secret-santa.fly.dev
PORT               8080
ENV                production | development
LOG_LEVEL          info
```

В dev подгружается из `.env` через `godotenv`, в проде — через `fly secrets`.

## 13. Безопасность

- **Session cookie:** `HttpOnly`, `Secure`, `SameSite=Lax`, имя `s`, токен 32 байта base64url, 30 дней.
- **Magic-link токены:** одноразовые (`used_at`), 15 минут, `crypto/rand`.
- **CSRF:** `SameSite=Lax` + все мутации идут как JSON POST — этого достаточно для MVP.
- **XSS:** React экранирует по умолчанию, `dangerouslySetInnerHTML` не используем. Wishlist и сообщения рендерятся как текст.
- **Invite-коды:** 12 символов из `a-z0-9` (~60 бит энтропии). Не секрет, но угадывание перебором нереально.
- **Authz:** каждый endpoint проверяет, что user — член группы / организатор до выдачи данных.
- **WebSocket origin check:** handshake проверяет `Origin` против `BASE_URL`.
- **Токены в логах** — только по префиксу (`abc...`).

## 14. Деплой

### Dockerfile (скелет)

```dockerfile
# Stage 1: build frontend
FROM node:20-alpine AS web
WORKDIR /web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: build Go binary with embedded frontend
FROM golang:1.26-alpine AS server
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

### fly.toml (ключевые моменты)

- Один регион (`fra` или `ams`).
- Один `shared-cpu-1x` инстанс (SQLite требует одного писателя).
- `[[mounts]]` volume `data` на `/data` (1 GB).
- `[http_service]`: `force_https=true`, `auto_stop_machines=true` (спит без трафика).
