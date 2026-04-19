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
	Type string `json:"type"`
	Role string `json:"role"`
	Body string `json:"body"`
}

type OutboundMessage struct {
	Type      string `json:"type"`
	ID        int64  `json:"id,omitempty"`
	Role      string `json:"role,omitempty"`
	FromMe    bool   `json:"from_me,omitempty"`
	Body      string `json:"body,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

type ClientMessage struct {
	Client  *Client
	Message InboundMessage
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
