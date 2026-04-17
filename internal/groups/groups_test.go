package groups_test

import (
	"regexp"
	"testing"

	"github.com/andreypisarev/secret-santa/internal/groups"
)

func TestGenerateInviteCode(t *testing.T) {
	code, err := groups.GenerateInviteCode()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(code) != 12 {
		t.Errorf("length = %d, want 12", len(code))
	}

	if !regexp.MustCompile(`^[a-z0-9]{12}$`).MatchString(code) {
		t.Errorf("code %q doesn't match [a-z0-9]{12}", code)
	}

	code2, _ := groups.GenerateInviteCode()
	if code == code2 {
		t.Error("two codes should not be equal")
	}
}
