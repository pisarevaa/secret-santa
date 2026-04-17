package email_test

import (
	"testing"

	"github.com/andreypisarev/secret-santa/internal/email"
)

func TestLogSender_Send(t *testing.T) {
	s := &email.LogSender{}
	err := s.Send("test@example.com", "Subject", "<p>Body</p>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
