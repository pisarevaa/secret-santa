package chat_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/andreypisarev/secret-santa/internal/chat"
	"github.com/andreypisarev/secret-santa/internal/db"
	"github.com/andreypisarev/secret-santa/internal/db/sqlc"
)

func sqlNullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: true}
}

type chatFixture struct {
	queries     *sqlc.Queries
	groupID     int64
	memberships map[int64]*chat.Membership
	users       map[string]int64 // "alice" | "bob" | "carol" → userID
}

func setupChatTest(t *testing.T) *chatFixture {
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
	ctx := context.Background()

	u1, _ := queries.CreateUser(ctx, sqlc.CreateUserParams{Email: "a@test.com", Name: "Alice"})
	u2, _ := queries.CreateUser(ctx, sqlc.CreateUserParams{Email: "b@test.com", Name: "Bob"})
	u3, _ := queries.CreateUser(ctx, sqlc.CreateUserParams{Email: "c@test.com", Name: "Carol"})

	group, _ := queries.CreateGroup(ctx, sqlc.CreateGroupParams{
		InviteCode: "chattest1234", Title: "Chat Test", OrganizerID: u1.ID,
	})

	memberships := map[int64]*chat.Membership{
		u1.ID: {UserID: u1.ID, RecipientID: u2.ID, SantaID: u3.ID},
		u2.ID: {UserID: u2.ID, RecipientID: u3.ID, SantaID: u1.ID},
		u3.ID: {UserID: u3.ID, RecipientID: u1.ID, SantaID: u2.ID},
	}

	return &chatFixture{
		queries:     queries,
		groupID:     group.ID,
		memberships: memberships,
		users:       map[string]int64{"alice": u1.ID, "bob": u2.ID, "carol": u3.ID},
	}
}

func readOutbound(t *testing.T, ch <-chan []byte) chat.OutboundMessage {
	t.Helper()
	select {
	case data := <-ch:
		var msg chat.OutboundMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		return msg
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for outbound message")
	}
	return chat.OutboundMessage{}
}

func TestHub_MessageDelivery(t *testing.T) {
	f := setupChatTest(t)

	hub := chat.NewHub(f.groupID, f.queries, f.memberships)
	go hub.Run()
	defer hub.Stop()

	alice := &chat.Client{
		UserID:  f.users["alice"],
		GroupID: f.groupID,
		Send:    make(chan []byte, 10),
	}
	bob := &chat.Client{
		UserID:  f.users["bob"],
		GroupID: f.groupID,
		Send:    make(chan []byte, 10),
	}

	hub.Register(alice)
	hub.Register(bob)

	hub.Incoming() <- chat.ClientMessage{
		Client: alice,
		Message: chat.InboundMessage{
			Type: "send",
			Role: "santa",
			Body: "Привет, подопечный!",
		},
	}

	aliceMsg := readOutbound(t, alice.Send)
	if aliceMsg.Type != "message" || !aliceMsg.FromMe || aliceMsg.Role != "santa" {
		t.Errorf("alice got unexpected: %+v", aliceMsg)
	}
	if aliceMsg.Body != "Привет, подопечный!" {
		t.Errorf("alice body = %q", aliceMsg.Body)
	}

	bobMsg := readOutbound(t, bob.Send)
	if bobMsg.Type != "message" || bobMsg.FromMe || bobMsg.Role != "recipient" {
		t.Errorf("bob got unexpected: %+v", bobMsg)
	}
	if bobMsg.Body != "Привет, подопечный!" {
		t.Errorf("bob body = %q", bobMsg.Body)
	}
}

func TestHub_RecipientRoleRoutesToSanta(t *testing.T) {
	f := setupChatTest(t)

	hub := chat.NewHub(f.groupID, f.queries, f.memberships)
	go hub.Run()
	defer hub.Stop()

	// Alice пишет как подопечный — сообщение должно прийти ее Санте (Carol).
	alice := &chat.Client{UserID: f.users["alice"], GroupID: f.groupID, Send: make(chan []byte, 10)}
	carol := &chat.Client{UserID: f.users["carol"], GroupID: f.groupID, Send: make(chan []byte, 10)}
	hub.Register(alice)
	hub.Register(carol)

	hub.Incoming() <- chat.ClientMessage{
		Client: alice,
		Message: chat.InboundMessage{Type: "send", Role: "recipient", Body: "Спасибо, Санта!"},
	}

	aliceMsg := readOutbound(t, alice.Send)
	if !aliceMsg.FromMe || aliceMsg.Role != "recipient" {
		t.Errorf("alice got unexpected: %+v", aliceMsg)
	}

	carolMsg := readOutbound(t, carol.Send)
	if carolMsg.FromMe || carolMsg.Role != "santa" {
		t.Errorf("carol got unexpected: %+v", carolMsg)
	}
	if carolMsg.Body != "Спасибо, Санта!" {
		t.Errorf("carol body = %q", carolMsg.Body)
	}
}

func TestHub_RateLimit(t *testing.T) {
	f := setupChatTest(t)

	hub := chat.NewHub(f.groupID, f.queries, f.memberships)
	go hub.Run()
	defer hub.Stop()

	client := &chat.Client{
		UserID:  f.users["alice"],
		GroupID: f.groupID,
		Send:    make(chan []byte, 64),
	}
	hub.Register(client)

	for range 11 {
		hub.Incoming() <- chat.ClientMessage{
			Client:  client,
			Message: chat.InboundMessage{Type: "send", Role: "santa", Body: "msg"},
		}
	}

	time.Sleep(200 * time.Millisecond)

	errorCount := 0
	messageCount := 0
drain:
	for {
		select {
		case data := <-client.Send:
			var msg chat.OutboundMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			switch msg.Type {
			case "error":
				errorCount++
			case "message":
				messageCount++
			}
		default:
			break drain
		}
	}
	if errorCount == 0 {
		t.Error("expected at least one rate limit error")
	}
	if messageCount > 10 {
		t.Errorf("messageCount = %d, want <= 10", messageCount)
	}
}

func TestHub_InvalidMessages(t *testing.T) {
	f := setupChatTest(t)

	hub := chat.NewHub(f.groupID, f.queries, f.memberships)
	go hub.Run()
	defer hub.Stop()

	client := &chat.Client{UserID: f.users["alice"], GroupID: f.groupID, Send: make(chan []byte, 10)}
	hub.Register(client)

	// Пустое тело
	hub.Incoming() <- chat.ClientMessage{
		Client:  client,
		Message: chat.InboundMessage{Type: "send", Role: "santa", Body: ""},
	}
	if msg := readOutbound(t, client.Send); msg.Type != "error" {
		t.Errorf("empty body: got %+v, want error", msg)
	}

	// Неверная роль
	hub.Incoming() <- chat.ClientMessage{
		Client:  client,
		Message: chat.InboundMessage{Type: "send", Role: "weird", Body: "hi"},
	}
	if msg := readOutbound(t, client.Send); msg.Type != "error" {
		t.Errorf("bad role: got %+v, want error", msg)
	}

	// Неизвестный тип
	hub.Incoming() <- chat.ClientMessage{
		Client:  client,
		Message: chat.InboundMessage{Type: "ping", Role: "santa", Body: "hi"},
	}
	if msg := readOutbound(t, client.Send); msg.Type != "error" {
		t.Errorf("bad type: got %+v, want error", msg)
	}
}

func TestHubManager_LoadsMembershipsFromDB(t *testing.T) {
	f := setupChatTest(t)
	ctx := context.Background()

	for _, name := range []string{"alice", "bob", "carol"} {
		_, err := f.queries.CreateMembership(ctx, sqlc.CreateMembershipParams{
			GroupID: f.groupID, UserID: f.users[name], Wishlist: "",
		})
		if err != nil {
			t.Fatalf("create membership %s: %v", name, err)
		}
	}

	// Выставить подопечных в БД по циклу Alice→Bob→Carol→Alice.
	_ = f.queries.SetRecipient(ctx, sqlc.SetRecipientParams{
		RecipientID: sqlNullInt64(f.users["bob"]), GroupID: f.groupID, UserID: f.users["alice"],
	})
	_ = f.queries.SetRecipient(ctx, sqlc.SetRecipientParams{
		RecipientID: sqlNullInt64(f.users["carol"]), GroupID: f.groupID, UserID: f.users["bob"],
	})
	_ = f.queries.SetRecipient(ctx, sqlc.SetRecipientParams{
		RecipientID: sqlNullInt64(f.users["alice"]), GroupID: f.groupID, UserID: f.users["carol"],
	})

	mgr := chat.NewHubManager(f.queries, nil)
	defer mgr.CloseAll()

	hub, err := mgr.GetOrCreateHub(f.groupID)
	if err != nil {
		t.Fatalf("get hub: %v", err)
	}

	alice := &chat.Client{UserID: f.users["alice"], GroupID: f.groupID, Send: make(chan []byte, 10)}
	bob := &chat.Client{UserID: f.users["bob"], GroupID: f.groupID, Send: make(chan []byte, 10)}
	hub.Register(alice)
	hub.Register(bob)

	hub.Incoming() <- chat.ClientMessage{
		Client:  alice,
		Message: chat.InboundMessage{Type: "send", Role: "santa", Body: "hi bob"},
	}

	aliceMsg := readOutbound(t, alice.Send)
	if !aliceMsg.FromMe {
		t.Errorf("alice: %+v", aliceMsg)
	}
	bobMsg := readOutbound(t, bob.Send)
	if bobMsg.FromMe || bobMsg.Body != "hi bob" {
		t.Errorf("bob: %+v", bobMsg)
	}
}
