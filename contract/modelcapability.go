package contract

import _ "embed"

//go:embed testdata/model_snapshot.json
var modelSnapshotJSON []byte

//go:embed testdata/models_meta.json
var modelsMetaJSON []byte

// ModelSnapshotJSON returns a copy of the model snapshot golden JSON payload.
func ModelSnapshotJSON() []byte {
	return append([]byte(nil), modelSnapshotJSON...)
}

// ModelsMetaJSON returns a copy of the models_meta golden JSON payload.
func ModelsMetaJSON() []byte {
	return append([]byte(nil), modelsMetaJSON...)
}
