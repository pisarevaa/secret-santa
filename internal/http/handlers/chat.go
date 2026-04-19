package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

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

	acceptOptions := &websocket.AcceptOptions{}
	if h.Config.IsDev() {
		acceptOptions.InsecureSkipVerify = true
	} else if h.Config.BaseURL != "" {
		acceptOptions.OriginPatterns = []string{h.Config.BaseURL}
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

	membership, err := h.Queries.GetMembershipByGroupAndUser(r.Context(), sqlc.GetMembershipByGroupAndUserParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		writeError(w, http.StatusForbidden, "forbidden", "вы не участник группы")
		return
	}

	var peerID int64

	if role == "santa" {
		if !membership.RecipientID.Valid {
			writeError(w, http.StatusNotFound, "not_found", "жеребьевка не проведена")
			return
		}
		peerID = membership.RecipientID.Int64
	} else {
		members, err := h.Queries.ListMembershipsByGroup(r.Context(), groupID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
			return
		}
		var santaID int64
		for _, m := range members {
			if m.RecipientID.Valid && m.RecipientID.Int64 == userID {
				santaID = m.UserID
				break
			}
		}
		if santaID == 0 {
			writeError(w, http.StatusNotFound, "not_found", "Санта не найден")
			return
		}
		peerID = santaID
	}

	msgs, err := h.Queries.ListChatMessages(r.Context(), sqlc.ListChatMessagesParams{
		GroupID:       groupID,
		SenderID:      userID,
		RecipientID:   peerID,
		SenderID_2:    peerID,
		RecipientID_2: userID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "внутренняя ошибка")
		return
	}

	out := make([]map[string]interface{}, 0, len(msgs))
	for _, m := range msgs {
		fromMe := m.SenderID == userID
		var msgRole string
		if role == "santa" {
			if fromMe {
				msgRole = "santa"
			} else {
				msgRole = "recipient"
			}
		} else {
			if fromMe {
				msgRole = "recipient"
			} else {
				msgRole = "santa"
			}
		}
		out = append(out, map[string]interface{}{
			"id":         m.ID,
			"role":       msgRole,
			"from_me":    fromMe,
			"body":       m.Body,
			"created_at": m.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, out)
}
