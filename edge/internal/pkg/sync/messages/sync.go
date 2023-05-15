package messages

import "devais.it/kronos/internal/pkg/types"

type SyncAction string

const (
	SyncActionCreate SyncAction = "CREATE"
	SyncActionUpdate SyncAction = "UPDATE"
	SyncActionDelete SyncAction = "DELETE"
)

type SyncEntry struct {
	EntityType types.EntityType       `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Version    string                 `json:"version"`
	Action     SyncAction             `json:"action"`
	Payload    map[string]interface{} `json:"payload"`
}

type Sync []SyncEntry
