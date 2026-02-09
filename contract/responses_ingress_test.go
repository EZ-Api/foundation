package contract

import (
	"slices"
	"testing"
)

func TestResponsesRequestFieldDecision(t *testing.T) {
	t.Parallel()

	if got := ResponsesRequestFieldDecision("model"); got != ResponsesFieldDecisionPassThrough {
		t.Fatalf("expected model to be pass-through, got %q", got)
	}
	if got := ResponsesRequestFieldDecision("reasoning"); got != ResponsesFieldDecisionReject {
		t.Fatalf("expected reasoning to be reject, got %q", got)
	}
	if got := ResponsesRequestFieldDecision("x_extension"); got != ResponsesFieldDecisionDegrade {
		t.Fatalf("expected extension to be degrade, got %q", got)
	}
}

func TestValidateResponsesRequestFields(t *testing.T) {
	t.Parallel()

	okPayload := map[string]any{
		"model":   "gpt-4.1",
		"input":   "hello",
		"unknown": map[string]any{"x": 1},
	}
	if err := ValidateResponsesRequestFields(okPayload); err != nil {
		t.Fatalf("expected payload to pass, got err=%v", err)
	}

	badPayload := map[string]any{
		"model":     "gpt-4.1",
		"reasoning": map[string]any{"effort": "high"},
	}
	if err := ValidateResponsesRequestFields(badPayload); err == nil {
		t.Fatalf("expected validation error for rejected field")
	}
}

func TestResponsesRequestFieldsList(t *testing.T) {
	t.Parallel()

	pass := ResponsesPassThroughRequestFields()
	reject := ResponsesRejectedRequestFields()

	if !slices.Contains(pass, "model") || !slices.Contains(pass, "metadata") {
		t.Fatalf("missing pass-through fields: %v", pass)
	}
	if !slices.Contains(reject, "reasoning") || !slices.Contains(reject, "previous_response_id") {
		t.Fatalf("missing rejected fields: %v", reject)
	}
}
