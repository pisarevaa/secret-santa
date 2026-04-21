# Ревью реализации: отчет

Дата ревью: 2026-04-21
Базовая ветка: `main` (HEAD `93341eb`)
Проверены фичи 1-10 из `docs/project-implementation.md`.

## Резюме

- Валидация прошла успешно: `go vet ./...`, `go build ./...`, `go test ./...` (все пакеты OK), `web: npx tsc --noEmit`, `web: npm run build` — без ошибок.
- Все 10 фич реализованы и соответствуют спецификации в ключевых аспектах (auth, groups, draw, chat, фронт, деплой, CI).
- Critical-проблем, ломающих сборку/тесты, не найдено.
- Найдена 1 Important-проблема (docker-compose по умолчанию молча ломает логин) и 7 Minor-проблем (мертвый код, плановый дрейф, N+1, маскирование ошибок).

## Классификация

- **Critical**: 0
- **Important**: 1 (исправлено: 1, осталось: 0)
- **Minor**: 7 (исправлено: 7, осталось: 0)

**Все проблемы из отчета исправлены (2026-04-21).**

---

## Important

### I-1. docker-compose.yml по умолчанию ломает логин — ИСПРАВЛЕНО (2026-04-21)

**Проблема.** `docker-compose.yml` задает `ENV: production`, но переменные `RESEND_API_KEY` и `EMAIL_FROM` закомментированы. В проде приложение выбирает `ResendSender` (`cmd/server/main.go:61-69`), а он обращается к Resend API с пустым `Authorization: Bearer ` — запрос падает с 401, но в `AuthHandler.RequestLink` ошибка отправки письма только логируется (`internal/http/handlers/auth.go:62-64`), клиенту возвращается 204. Пользователь, запустивший `docker compose up` без правки файла, не получает magic-link и не может залогиниться. Плюс `Config.Load` (`internal/config/config.go:19-56`) не валидирует обязательные переменные в production-режиме.

**Файлы.** `docker-compose.yml:8-14`, `internal/config/config.go:47-55`, `internal/email/email.go:30-58`.

**Шаги исправления.**
- [x] Поменять дефолт в `docker-compose.yml` на `ENV: development` (локальный `docker compose up` тогда использует `LogSender` и не требует Resend).
- [x] Добавить в `config.Load()` проверку, что при `ENV=production` заданы непустые `RESEND_API_KEY` и `EMAIL_FROM`, иначе вернуть ошибку при старте.
- [x] Добавить в `ResendSender.Send` ранний возврат ошибки, если `s.APIKey == ""` или `s.From == ""` — это даст явную диагностику вместо 401 от Resend.

**Проверка.**
```bash
docker compose up -d && sleep 3
curl -s -o /dev/null -w "%{http_code}\n" -X POST http://localhost:8080/api/auth/request-link \
  -H 'Content-Type: application/json' -d '{"email":"test@example.com"}'
docker compose logs app | grep -i email
```
Ожидание: либо лог письма (`LogSender`) в dev-режиме, либо приложение падает при старте с понятной ошибкой про отсутствующие переменные.

---

## Minor

### M-1. Плановый дрейф: Go 1.22 → 1.26 — ИСПРАВЛЕНО (2026-04-21)

**Проблема.** План (`docs/project-implementation.md:4542`, `:4638`) закреплен на Go 1.22. Фактически `go.mod` требует `go 1.26.1`, `Dockerfile:10` — `golang:1.26-alpine`, `.github/workflows/ci.yml:17` — `go-version: "1.26"`. Go 1.26 существует (проверено: `go version` даёт `go1.26.1`), сборка и тесты зеленые — то есть это сознательный апгрейд, а не баг. Но расхождение с планом не задокументировано.

**Файлы.** `go.mod:3`, `Dockerfile:10`, `.github/workflows/ci.yml:17`, `docs/project-implementation.md:4542,4638`.

**Шаги исправления.**
- [x] Обновить план, указав актуальную версию Go (1.26): `docs/project-implementation.md` (строки 9, 4542, 4638) и `docs/project-design.md` (строки 29, 346).

**Проверка.**
```bash
go version
grep -E "go 1\.|golang:1\.|go-version" go.mod Dockerfile .github/workflows/ci.yml
```
Ожидание: все версии согласованы между собой и с планом.

### M-2. Мертвый код: `Hub.BroadcastDrawn` — ИСПРАВЛЕНО (2026-04-21)

**Проблема.** Метод `Hub.BroadcastDrawn` (`internal/chat/hub.go:192-202`) нигде не вызывается из прод-кода — только из теста `internal/chat/hub_test.go:221`. В `DrawHandler.Draw` (`internal/http/handlers/draw.go:22-110`) после коммита транзакции нет оповещения хаба. К тому же метод читает `h.clients` без синхронизации с Run-горутиной, которая владеет этой картой — это латентная гонка данных, если когда-нибудь метод начнут вызывать из другой горутины.

**Файлы.** `internal/chat/hub.go:192-202`, `internal/http/handlers/draw.go`.

**Шаги исправления.**
- [x] Удалить `BroadcastDrawn` и соответствующий тест (YAGNI — фронт и так перезагружает страницу после draw).

**Проверка.**
```bash
grep -rn "BroadcastDrawn" --include='*.go' .
go test ./internal/chat/... -race -v
```
Ожидание: либо метод удалён и тесты проходят, либо вызов появляется в draw handler и `-race` не находит гонок.

### M-3. YAGNI: неиспользуемые SQL-запросы в messages.sql — ИСПРАВЛЕНО (2026-04-21)

**Проблема.** Запросы `ListMessages` и `ListMessagesBefore` (`internal/db/queries/messages.sql:4-14`) сгенерированы sqlc (`internal/db/sqlc/messages.sql.go:98-189`), но нигде не вызываются из Go-кода. Используется только `ListChatMessages`.

**Файлы.** `internal/db/queries/messages.sql:4-14`, `internal/db/sqlc/messages.sql.go:98-189`.

**Шаги исправления.**
- [x] Удалить `ListMessages` и `ListMessagesBefore` из `messages.sql`.
- [x] Перегенерировать sqlc: `sqlc generate`.
- [x] Убедиться, что сборка зеленая.

**Проверка.**
```bash
grep -rn "ListMessages\b\|ListMessagesBefore" --include='*.go' .
go build ./...
go test ./...
```
Ожидание: grep находит только `ListChatMessages`, сборка и тесты зеленые.

### M-4. N+1 запрос в `GroupHandler.GetByInviteCode` — ИСПРАВЛЕНО (2026-04-21)

**Проблема.** В `internal/http/handlers/groups.go:95-107` для каждого участника группы выполняется отдельный `GetUserByID`. При группе на N человек это N+1 запрос. Для нынешнего размера групп не критично, но противоречит принципу «одной SQL-ходки».

**Файлы.** `internal/http/handlers/groups.go:85-107`, `internal/db/queries/memberships.sql`.

**Шаги исправления.**
- [x] Добавить в `memberships.sql` запрос `ListMembersWithNamesByGroup` с `JOIN users ON users.id = memberships.user_id`, возвращающий `id, user_id, name`.
- [x] Перегенерировать sqlc и заменить цикл с `GetUserByID` на один вызов.

**Проверка.**
```bash
go test ./internal/http/handlers/... -run TestSmoke -v
```
Ожидание: smoke-тест проходит, ответ `GET /api/groups/{invite}` не меняет форму.

### M-5. Проглатывается ошибка `DeleteSession` в Logout — ИСПРАВЛЕНО (2026-04-21)

**Проблема.** `internal/http/handlers/auth.go:140` использует `_ = h.Queries.DeleteSession(...)`. Ошибка БД не логируется — при проблеме с БД сессия клиента будет помечена на удаление cookie, но серверная запись останется активной, а разработчик не увидит этого в логах.

**Файлы.** `internal/http/handlers/auth.go:133-151`.

**Шаги исправления.**
- [x] Заменить `_ =` на `if err := ...; err != nil { slog.Error("delete session", "error", err) }`.

**Проверка.**
```bash
go vet ./...
go test ./internal/http/handlers/... -v
```
Ожидание: vet/тесты зеленые.

### M-6. `GetMagicLink` маскирует ошибки БД как 400 — ИСПРАВЛЕНО (2026-04-21)

**Проблема.** В `internal/http/handlers/auth.go:77-81` любая ошибка `GetMagicLink` (включая ошибку соединения с БД) возвращается пользователю как 400 «ссылка недействительна или истекла». Реальные 5xx-ошибки маскируются под 4xx, и их сложно диагностировать в мониторинге.

**Файлы.** `internal/http/handlers/auth.go:77-81`.

**Шаги исправления.**
- [x] Различать `sql.ErrNoRows` (реально невалидный/истекший токен → 400) и прочие ошибки (→ 500 + `slog.Error`).

**Проверка.**
```bash
go test ./internal/http/handlers/... -v
```
Ожидание: существующие тесты проходят, ручное тестирование при сломанном пути БД даёт 500 с логом.

### M-7. `ResendSender.Send` не валидирует APIKey/From — ИСПРАВЛЕНО (2026-04-21, в рамках I-1)

**Проблема.** `internal/email/email.go:30-58` не проверяет, что `s.APIKey` и `s.From` непустые. При пустых значениях запрос всё равно уходит в Resend, возвращая 401 — сообщение об ошибке становится менее информативным. Связано с I-1.

**Файлы.** `internal/email/email.go:30-58`.

**Шаги исправления.**
- [x] В начале `Send` добавить проверку: `if s.APIKey == "" || s.From == "" { return errors.New("resend: API key or sender not configured") }`.

**Проверка.**
```bash
go test ./internal/email/... -v
```
Ожидание: добавить тест на пустые параметры, он должен проходить; существующие тесты — не ломаются.

---

## Что проверено и соответствует плану

- **Фича 1 (init)**: `go.mod`, структура директорий, `cmd/server/main.go` — OK.
- **Фича 2 (DB + sqlc)**: миграции в `internal/db/migrations/`, sqlc-генерация, WAL + `SetMaxOpenConns(1)` для сериализации записей — OK.
- **Фича 3 (auth)**: magic link 15 мин, сессия 30 дней, HttpOnly cookie, `Secure` по `IsDev()`, 204 на `RequestLink` (анти-энумерация), `errors.Is(sql.ErrNoRows)`, корректная сериализация `time.Time` через sqlc-типы — OK. Отступление от плана в пользу sqlc-типов вместо `.Format(time.DateTime)` оправдано.
- **Фича 4 (groups)**: проверки длины (utf8.RuneCount), запрет join после draw, проверка `membership.UserID == userID` на UpdateWishlist, обработка ошибки `UpdateUserName` (улучшение относительно плана) — OK.
- **Фича 5 (draw)**: алгоритм `perm[i] → perm[(i+1) % n]` в `internal/draw/`, транзакция BeginTx → `DrawGroup` (с проверкой `RowsAffected == 0` для идемпотентности) → `SetRecipient` в цикле → Commit, проверка организатора и статуса — OK. `math/rand` приемлем для негласных игр; план допускает.
- **Фича 6 (chat WebSocket)**: hub-per-group, register/unregister/incoming каналы, rate-limit 10 msg/min через переиспользование слайса (`filtered := h.rateCount[userID][:0]`), `MembershipByUserID` загружается при создании хаба и не меняется (валидно — после draw состав группы не меняется) — OK. Ограничение 2-person-group (где `ListChatMessages` возвращает и Santa→Recipient, и Recipient→Santa в обоих чатах) задокументировано в тесте и не противоречит плану.
- **Фича 6 (OriginPatterns)**: `internal/http/handlers/chat.go:59-61` задает `OriginPatterns = []string{h.Config.BaseURL}` — формально это URL со схемой, а `coder/websocket` ждет host-паттерн. Но для same-origin запросов `authenticateOrigin` проходит раньше, по совпадению `r.Host` и `u.Host` (`/Users/andreypisarev/go/pkg/mod/github.com/coder/websocket@v1.8.14/accept.go:239`). То есть в реальной проде (браузер → тот же домен) настройка никогда не используется, и запросы не отклоняются. Код избыточен, но не сломан.
- **Фича 7 (frontend)**: `createBrowserRouter` с 3 маршрутами, `onLogin={() => window.location.reload()}`, `GroupPage.tsx` извлекает `group.id` из API (закрыт TODO из плана) — OK.
- **Фича 8 (static + smoke-тест)**: `embed.FS`, SPA-фоллбек с исключением `/api/`, `/ws/`, `/healthz`, smoke-тест с реальной magic-link флоу через парсинг HTML письма — OK.
- **Фича 9 (Dockerfile, fly.toml)**: multi-stage build (node → golang → alpine), non-root `app` пользователь, `adduser -D -u 1000`, `chown app:app /data`, `fly.toml` с mounts — OK.
- **Фича 10 (CI)**: GitHub Actions с Go-тестами, TS-typecheck, frontend-сборкой — OK.

## Команды проверки

```bash
go vet ./...
go build ./...
go test ./...
cd web && npx tsc --noEmit && npm run build
```

Все зеленые на момент ревью (2026-04-21, HEAD `93341eb`).
