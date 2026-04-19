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

type Membership struct {
	UserID      int64
	RecipientID int64
	SantaID     int64
}

type Hub struct {
	groupID     int64
	clients     map[int64]map[*Client]struct{}
	memberships map[int64]*Membership
	register    chan *Client
	unregister  chan *Client
	incoming    chan ClientMessage
	quit        chan struct{}
	queries     *sqlc.Queries

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
				if _, exists := conns[client]; exists {
					delete(conns, client)
					close(client.Send)
				}
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
		if membership.RecipientID == 0 {
			h.sendError(sender, "подопечный не назначен")
			return
		}
		dbSenderID = sender.UserID
		dbRecipientID = membership.RecipientID
		direction = "santa_to_recipient"
	} else {
		if membership.SantaID == 0 {
			h.sendError(sender, "Санта не назначен")
			return
		}
		dbSenderID = sender.UserID
		dbRecipientID = membership.SantaID
		direction = "recipient_to_santa"
	}

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

	createdAt := saved.CreatedAt.UTC().Format(time.RFC3339)

	h.sendToUser(sender.UserID, OutboundMessage{
		Type:      "message",
		ID:        saved.ID,
		Role:      msg.Role,
		FromMe:    true,
		Body:      msg.Body,
		CreatedAt: createdAt,
	})

	var recipientRole string
	if direction == "santa_to_recipient" {
		recipientRole = "recipient"
	} else {
		recipientRole = "santa"
	}

	h.sendToUser(dbRecipientID, OutboundMessage{
		Type:      "message",
		ID:        saved.ID,
		Role:      recipientRole,
		FromMe:    false,
		Body:      msg.Body,
		CreatedAt: createdAt,
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

	members, err := m.queries.ListMembershipsByGroup(context.Background(), groupID)
	if err != nil {
		return nil, err
	}

	recipientOf := make(map[int64]int64)
	for _, mem := range members {
		if mem.RecipientID.Valid {
			recipientOf[mem.UserID] = mem.RecipientID.Int64
		}
	}

	santaOf := make(map[int64]int64)
	for santa, recipient := range recipientOf {
		santaOf[recipient] = santa
	}

	memberships := make(map[int64]*Membership)
	for _, mem := range members {
		memberships[mem.UserID] = &Membership{
			UserID:      mem.UserID,
			RecipientID: recipientOf[mem.UserID],
			SantaID:     santaOf[mem.UserID],
		}
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
