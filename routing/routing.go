package routing

import (
	"fmt"
	"regexp"
	"strings"
)

// ModelRef is a parsed representation of a client-facing model identifier.
// The canonical format is "namespace.public_model".
type ModelRef struct {
	Namespace   string
	PublicModel string
}

func (m ModelRef) Key() string {
	if strings.TrimSpace(m.Namespace) == "" || strings.TrimSpace(m.PublicModel) == "" {
		return ""
	}
	return strings.TrimSpace(m.Namespace) + "." + strings.TrimSpace(m.PublicModel)
}

// ParseModelRef parses client-provided model string.
// If model contains '.', it is treated as "namespace.public_model" (split on first dot).
// Otherwise, defaultNamespace is used as namespace.
func ParseModelRef(model string, defaultNamespace string) (ModelRef, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		return ModelRef{}, fmt.Errorf("model required")
	}

	if ns, rest, ok := strings.Cut(model, "."); ok {
		ns = strings.TrimSpace(ns)
		rest = strings.TrimSpace(rest)
		if ns == "" || rest == "" {
			return ModelRef{}, fmt.Errorf("invalid model: %q", model)
		}
		return ModelRef{Namespace: ns, PublicModel: rest}, nil
	}

	defaultNamespace = strings.TrimSpace(defaultNamespace)
	if defaultNamespace == "" {
		return ModelRef{}, fmt.Errorf("default namespace required")
	}
	return ModelRef{Namespace: defaultNamespace, PublicModel: model}, nil
}

func NormalizeModelID(id string) string {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		return ""
	}
	// Conservative normalization for common provider variants like "moonshot/kimi2".
	if i := strings.LastIndex(id, "/"); i >= 0 && i+1 < len(id) {
		id = id[i+1:]
	}
	return strings.TrimSpace(id)
}

type SelectorType string

const (
	SelectorExact          SelectorType = "exact"
	SelectorRegex          SelectorType = "regex"
	SelectorNormalizeExact SelectorType = "normalize_exact"
)

// ResolveUpstreamModel resolves a single upstream model name for a provider given a selector.
// It enforces the "unique hit" rule: 0 hit or >1 hit is an error.
func ResolveUpstreamModel(selectorType SelectorType, selectorValue string, publicModel string, providerModels []string) (string, error) {
	v := strings.TrimSpace(selectorValue)
	if v == "" {
		v = strings.TrimSpace(publicModel)
	}
	if v == "" {
		return "", fmt.Errorf("selector value missing")
	}

	switch selectorType {
	case "", SelectorExact:
		for _, m := range providerModels {
			if strings.TrimSpace(m) == v {
				return strings.TrimSpace(m), nil
			}
		}
		return "", fmt.Errorf("no match for %q", v)
	case SelectorRegex:
		re, err := regexp.Compile(v)
		if err != nil {
			return "", fmt.Errorf("invalid regex: %w", err)
		}
		var hits []string
		for _, m := range providerModels {
			m2 := strings.TrimSpace(m)
			if m2 == "" {
				continue
			}
			if re.MatchString(m2) {
				hits = append(hits, m2)
			}
		}
		if len(hits) == 1 {
			return hits[0], nil
		}
		if len(hits) == 0 {
			return "", fmt.Errorf("no regex match for %q", v)
		}
		return "", fmt.Errorf("regex matched multiple models (%d)", len(hits))
	case SelectorNormalizeExact:
		want := NormalizeModelID(v)
		var hit string
		for _, m := range providerModels {
			m2 := strings.TrimSpace(m)
			if m2 == "" {
				continue
			}
			if NormalizeModelID(m2) == want {
				if hit != "" {
					return "", fmt.Errorf("normalize matched multiple models")
				}
				hit = m2
			}
		}
		if hit == "" {
			return "", fmt.Errorf("no normalize match for %q", v)
		}
		return hit, nil
	default:
		return "", fmt.Errorf("unsupported selector type: %q", string(selectorType))
	}
}

// BindingCandidate represents a single provider group candidate for a bindingKey.
type BindingCandidate struct {
	GroupID       uint              `json:"group_id"`
	RouteGroup    string            `json:"route_group"`
	Weight        int               `json:"weight,omitempty"`
	SelectorType  string            `json:"selector_type,omitempty"`
	SelectorValue string            `json:"selector_value,omitempty"`
	Status        string            `json:"status,omitempty"`
	Error         string            `json:"error,omitempty"` // config_error | no_provider
	Upstreams     map[string]string `json:"upstreams"`       // provider_id -> upstream_model
}

// BindingSnapshot is the DP-consumed snapshot for "(namespace, public_model) -> candidates -> provider -> upstream_model".
// DP hot path should only do O(1) map lookups.
type BindingSnapshot struct {
	Namespace   string             `json:"namespace"`
	PublicModel string             `json:"public_model"`
	Status      string             `json:"status,omitempty"`
	UpdatedAt   int64              `json:"updated_at,omitempty"` // unix seconds
	Candidates  []BindingCandidate `json:"candidates"`
}
