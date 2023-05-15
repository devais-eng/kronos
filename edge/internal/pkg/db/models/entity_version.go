package models

// EntityVersion is a structure representing the version of a model.
// This is returned by version queries.
type EntityVersion struct {
	ID          string `json:"id"`
	Version     string `json:"version"`
	SyncVersion string `json:"sync_version"`
	ModifiedAt  uint64 `json:"modified_at"`
	ModifiedBy  string `json:"modified_by"`
}
