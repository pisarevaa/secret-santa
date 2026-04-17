package groups

import (
	"crypto/rand"
	"math/big"
)

const inviteCodeAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
const inviteCodeLength = 12

// GenerateInviteCode создает случайный код из 12 символов [a-z0-9].
func GenerateInviteCode() (string, error) {
	code := make([]byte, inviteCodeLength)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(inviteCodeAlphabet))))
		if err != nil {
			return "", err
		}
		code[i] = inviteCodeAlphabet[n.Int64()]
	}
	return string(code), nil
}
