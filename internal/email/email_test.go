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

func TestResendSender_Send_MissingConfig(t *testing.T) {
	cases := []struct {
		name   string
		sender email.ResendSender
	}{
		{"empty APIKey", email.ResendSender{APIKey: "", From: "from@example.com"}},
		{"empty From", email.ResendSender{APIKey: "re_test", From: ""}},
		{"both empty", email.ResendSender{APIKey: "", From: ""}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.sender.Send("to@example.com", "Subject", "<p>Body</p>")
			if err == nil {
				t.Fatal("expected error for missing config, got nil")
			}
		})
	}
}
