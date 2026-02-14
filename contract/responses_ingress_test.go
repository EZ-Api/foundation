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
	if got := ResponsesRequestFieldDecision("reasoning"); got != ResponsesFieldDecisionPassThrough {
		t.Fatalf("expected reasoning to be pass-through, got %q", got)
	}
	if got := ResponsesRequestFieldDecision("messages"); got != ResponsesFieldDecisionReject {
		t.Fatalf("expected messages to be reject, got %q", got)
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
		"model":    "gpt-4.1",
		"messages": []any{},
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
	if !slices.Contains(pass, "previous_response_id") || !slices.Contains(pass, "reasoning") {
		t.Fatalf("missing pass-through responses native fields: %v", pass)
	}
	if !slices.Contains(reject, "messages") {
		t.Fatalf("missing rejected fields: %v", reject)
	}
}
