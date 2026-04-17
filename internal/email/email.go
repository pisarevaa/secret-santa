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
		"to":      []string{to},
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
