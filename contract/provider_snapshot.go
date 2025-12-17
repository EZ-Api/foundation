package contract

import _ "embed"

//go:embed testdata/provider_snapshot.json
var providerSnapshotJSON []byte

// ProviderSnapshotJSON returns a copy of the provider snapshot golden JSON payload.
func ProviderSnapshotJSON() []byte {
	return append([]byte(nil), providerSnapshotJSON...)
}
