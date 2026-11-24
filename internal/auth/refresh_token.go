package auth

import (
	"crypto/rand"
	"encoding/hex"
)

// Returns a 64 char long string of random characters.
// Represents the Hex encoding of randomly generated 32 bytes.
// Each byte has is 8 bits, 1 Hex char can represent 4 bits, so 1 byte = 2 hex char.
// Since there are 32 bits, there will be 32 * 2 = 64 characters in the Hex Encoding.
func MakeRefreshToken() string {
	token := make([]byte, 32)
	_, _ = rand.Read(token)
	encodedToken := hex.EncodeToString(token)
	return encodedToken
}
