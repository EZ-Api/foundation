package modelcap

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
)

type Kind string

const (
	KindChat      Kind = "chat"
	KindEmbedding Kind = "embedding"
	KindRerank    Kind = "rerank"
	KindOther     Kind = "other"
)

func NormalizeKind(kind string) Kind {
	k := strings.ToLower(strings.TrimSpace(kind))
	if k == "" {
		return KindChat
	}
	switch Kind(k) {
	case KindChat, KindEmbedding, KindRerank, KindOther:
		return Kind(k)
	default:
		return KindOther
	}
}

// Model is the canonical schema stored in Redis meta:models (hash value JSON).
// It is keyed by bindingKey (namespace.public_model).
type Model struct {
	Name               string  `json:"name"`
	Kind               string  `json:"kind,omitempty"`
	ContextWindow      int     `json:"context_window,omitempty"`
	CostPerToken       float64 `json:"cost_per_token,omitempty"`
	SupportsVision     bool    `json:"supports_vision,omitempty"`
	SupportsFunction   bool    `json:"supports_functions,omitempty"`
	SupportsToolChoice bool    `json:"supports_tool_choice,omitempty"`
	SupportsFim        bool    `json:"supports_fim,omitempty"`
	SupportsStream     bool    `json:"supports_stream,omitempty"`
	MaxOutputTokens    int     `json:"max_output_tokens,omitempty"`
}

func (m Model) Normalized() Model {
	m.Name = strings.TrimSpace(m.Name)
	m.Kind = string(NormalizeKind(m.Kind))
	return m
}

func (m Model) Validate() error {
	m = m.Normalized()
	if m.Name == "" {
		return errors.New("name required")
	}
	if m.ContextWindow < 0 {
		return errors.New("context_window must be >= 0")
	}
	if m.MaxOutputTokens < 0 {
		return errors.New("max_output_tokens must be >= 0")
	}
	return nil
}

// Meta is stored in Redis meta:models_meta (hash).
type Meta struct {
	Version     string `json:"version"`
	UpdatedAt   string `json:"updated_at"`
	Source      string `json:"source"`
	Checksum    string `json:"checksum"`
	UpstreamURL string `json:"upstream_url,omitempty"`
	UpstreamRef string `json:"upstream_ref,omitempty"`
}

func ChecksumFromPayloads(payloads map[string]string) string {
	keys := make([]string, 0, len(payloads))
	for k := range payloads {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		_, _ = h.Write([]byte(k))
		_, _ = h.Write([]byte{'\n'})
		_, _ = h.Write([]byte(payloads[k]))
		_, _ = h.Write([]byte{'\n'})
	}
	return hex.EncodeToString(h.Sum(nil))
}
