package messages

import "devais.it/kronos/internal/pkg/types"

type Event struct {
	ID          uint                   `json:"id"`
	EntityType  types.EntityType       `json:"entity_type"`
	EntityID    string                 `json:"entity_id"`
	TriggeredBy string                 `json:"triggered_by"`
	TxUUID      string                 `json:"tx_uuid,omitempty"`
	TxType      types.EventType        `json:"tx_type"`
	TxLen       int                    `json:"tx_len,omitempty"`
	TxIndex     int                    `json:"tx_index"`
	Timestamp   uint64                 `json:"timestamp"`
	Body        map[string]interface{} `json:"body,omitempty"`
}
