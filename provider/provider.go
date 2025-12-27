package provider

import "strings"

const (
	TypeOpenAI        = "openai"
	TypeCompatible    = "compatible"
	TypeAnthropic     = "anthropic"
	TypeClaude        = "claude"
	TypeClaudeCode    = "claude-code"
	TypeCodex         = "codex"
	TypeGeminiCLI     = "gemini-cli"
	TypeAntigravity   = "antigravity"
	TypeGemini        = "gemini"
	TypeGoogle        = "google"
	TypeAIStudio      = "aistudio"
	TypeVertex        = "vertex"
	TypeVertexExpress = "vertex-express"
)

var googleFamily = map[string]struct{}{
	TypeGemini:        {},
	TypeGoogle:        {},
	TypeAIStudio:      {},
	TypeVertex:        {},
	TypeVertexExpress: {},
}

func NormalizeType(t string) string {
	return strings.ToLower(strings.TrimSpace(t))
}

func IsGoogleFamily(providerType string) bool {
	_, ok := googleFamily[NormalizeType(providerType)]
	return ok
}

func IsVertexFamily(providerType string) bool {
	switch NormalizeType(providerType) {
	case TypeVertex, TypeVertexExpress:
		return true
	default:
		return false
	}
}

// DefaultGoogleLocation returns a normalized location, defaulting to "global" for Vertex providers.
func DefaultGoogleLocation(providerType, location string) string {
	location = strings.TrimSpace(location)
	if location == "" && IsVertexFamily(providerType) {
		return "global"
	}
	return location
}

// GoogleFamilyTypes returns provider types that should be handled by Google transport/channel/adapter.
func GoogleFamilyTypes() []string {
	return []string{TypeGemini, TypeGoogle, TypeAIStudio, TypeVertex, TypeVertexExpress}
}
