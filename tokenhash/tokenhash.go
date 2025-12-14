package tokenhash

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken returns sha256(token) in lowercase hex.
// This must remain stable because it is used as a cross-service contract (CP writes, DP validates).
func HashToken(token string) string {
	hasher := sha256.New()
	hasher.Write([]byte(token))
	return hex.EncodeToString(hasher.Sum(nil))
}

