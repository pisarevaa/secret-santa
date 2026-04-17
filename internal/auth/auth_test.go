package auth_test

import (
	"testing"

	"github.com/andreypisarev/secret-santa/internal/auth"
)

func TestGenerateToken(t *testing.T) {
	token1, err := auth.GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(token1) == 0 {
		t.Fatal("token is empty")
	}

	token2, err := auth.GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token1 == token2 {
		t.Error("two tokens should not be equal")
	}

	// 32 байта в base64url без паддинга = 43 символа
	if len(token1) != 43 {
		t.Errorf("token length = %d, want 43", len(token1))
	}
}
