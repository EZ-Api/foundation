package requestid

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

const HeaderName = "X-Request-ID"

// Extract returns request_id from headers (X-Request-ID / X-Request-Id), trimmed.
// The getter is typically http.Header.Get or gin.Context.GetHeader.
func Extract(get func(string) string) string {
	if get == nil {
		return ""
	}
	id := strings.TrimSpace(get(HeaderName))
	if id == "" {
		id = strings.TrimSpace(get("X-Request-Id"))
	}
	return id
}

// New generates a new request_id as lower hex.
func New() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(b[:])
}

