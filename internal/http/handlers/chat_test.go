package handlers_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/andreypisarev/secret-santa/internal/chat"
	"github.com/andreypisarev/secret-santa/internal/config"
	"github.com/andreypisarev/secret-santa/internal/db"
	"github.com/andreypisarev/secret-santa/internal/db/sqlc"
	"github.com/andreypisarev/secret-santa/internal/http/handlers"
	"github.com/go-chi/chi/v5"
)

func setupChat(t *testing.T) (*handlers.ChatHandler, *sqlc.Queries, *sql.DB) {
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
	hm := chat.NewHubManager(queries, database)
	t.Cleanup(hm.CloseAll)

	return &handlers.ChatHandler{
		Queries:    queries,
		HubManager: hm,
		Config:     &config.Config{Env: "development"},
	}, queries, database
}

func drawnGroup(t *testing.T, q *sqlc.Queries) (groupID, aliceID, bobID int64) {
	t.Helper()
	ctx := context.Background()
	alice, _ := q.CreateUser(ctx, sqlc.CreateUserParams{Email: "alice@test.com", Name: "Alice"})
	bob, _ := q.CreateUser(ctx, sqlc.CreateUserParams{Email: "bob@test.com", Name: "Bob"})

	group, _ := q.CreateGroup(ctx, sqlc.CreateGroupParams{
		InviteCode: "chathandler1", Title: "H", OrganizerID: alice.ID,
	})
	q.CreateMembership(ctx, sqlc.CreateMembershipParams{GroupID: group.ID, UserID: alice.ID, Wishlist: ""})
	q.CreateMembership(ctx, sqlc.CreateMembershipParams{GroupID: group.ID, UserID: bob.ID, Wishlist: ""})

	// Alice → Bob, Bob → Alice
	q.SetRecipient(ctx, sqlc.SetRecipientParams{
		RecipientID: sql.NullInt64{Int64: bob.ID, Valid: true},
		GroupID:     group.ID, UserID: alice.ID,
	})
	q.SetRecipient(ctx, sqlc.SetRecipientParams{
		RecipientID: sql.NullInt64{Int64: alice.ID, Valid: true},
		GroupID:     group.ID, UserID: bob.ID,
	})
	return group.ID, alice.ID, bob.ID
}

func TestChatHistory_SantaRole(t *testing.T) {
	h, q, _ := setupChat(t)
	ctx := context.Background()
	groupID, aliceID, bobID := drawnGroup(t, q)

	// Alice (Санта) → Bob (ее подопечный)
	q.CreateMessage(ctx, sqlc.CreateMessageParams{
		GroupID: groupID, SenderID: aliceID, RecipientID: bobID,
		Direction: "santa_to_recipient", Body: "Привет от Санты",
	})
	// Bob (подопечный) → Alice (его Санта)
	q.CreateMessage(ctx, sqlc.CreateMessageParams{
		GroupID: groupID, SenderID: bobID, RecipientID: aliceID,
		Direction: "recipient_to_santa", Body: "Привет, Санта",
	})

	r := chi.NewRouter()
	r.Get("/api/groups/{id}/chats/{role}", h.History)

	// Alice смотрит чат «как Санта» — здесь пара (Alice, Bob).
	// Ожидаем оба сообщения: Alice→Bob и неожиданно Bob→Alice также, потому что пара = {Alice, Bob}.
	// НО: направление "santa_to_recipient" между Alice/Bob — это Alice→Bob,
	// а "recipient_to_santa" Bob→Alice — это Bob пишет своему Санте Alice (другой диалог).
	// Тем не менее ListChatMessages выбирает обе пары (sender=me,recipient=peer) и (sender=peer,recipient=me).
	// Поэтому здесь в «Санта»-чате Alice увидит оба сообщения, что технически делает чат «Санта» и «подопечный» одной веткой для Alice↔Bob.
	req := httptest.NewRequest("GET", "/api/groups/"+strconv.FormatInt(groupID, 10)+"/chats/santa", nil)
	req = withUserID(req, aliceID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var msgs []map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&msgs); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("msgs count = %d, want 2", len(msgs))
	}

	// Первое (по ASC времени) — от Alice
	if fromMe, _ := msgs[0]["from_me"].(bool); !fromMe {
		t.Errorf("msgs[0] from_me = false, want true")
	}
	if role, _ := msgs[0]["role"].(string); role != "santa" {
		t.Errorf("msgs[0] role = %q, want santa", role)
	}
}

func TestChatHistory_NotMember(t *testing.T) {
	h, q, _ := setupChat(t)
	ctx := context.Background()
	groupID, _, _ := drawnGroup(t, q)

	outsider, _ := q.CreateUser(ctx, sqlc.CreateUserParams{Email: "out@test.com", Name: "O"})

	r := chi.NewRouter()
	r.Get("/api/groups/{id}/chats/{role}", h.History)

	req := httptest.NewRequest("GET", "/api/groups/"+strconv.FormatInt(groupID, 10)+"/chats/santa", nil)
	req = withUserID(req, outsider.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestChatHistory_InvalidRole(t *testing.T) {
	h, q, _ := setupChat(t)
	groupID, aliceID, _ := drawnGroup(t, q)

	r := chi.NewRouter()
	r.Get("/api/groups/{id}/chats/{role}", h.History)

	req := httptest.NewRequest("GET", "/api/groups/"+strconv.FormatInt(groupID, 10)+"/chats/elf", nil)
	req = withUserID(req, aliceID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
