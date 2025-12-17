package contract

import (
	"encoding/json"
	"testing"

	"github.com/ez-api/foundation/modelcap"
)

func TestModelSnapshotGolden_IsValid(t *testing.T) {
	var m modelcap.Model
	if err := json.Unmarshal(ModelSnapshotJSON(), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if m.Name == "" {
		t.Fatalf("expected name")
	}
}

func TestModelsMetaGolden_IsValid(t *testing.T) {
	var meta modelcap.Meta
	if err := json.Unmarshal(ModelsMetaJSON(), &meta); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if meta.Version == "" || meta.Source == "" || meta.UpdatedAt == "" || meta.Checksum == "" {
		t.Fatalf("expected required meta fields, got %+v", meta)
	}
}
