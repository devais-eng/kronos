package models

import (
	"devais.it/kronos/internal/pkg/types"
	jsoniter "github.com/json-iterator/go"
)

type Event struct {
	ID          uint             `gorm:"primaryKey" json:"id"`
	EventType   types.EventType  `gorm:"type:char(20);" json:"event_type"`
	EntityType  types.EntityType `gorm:"type:char(20);" json:"entity_type"`
	EntityID    string           `gorm:"type:char(128);" json:"entity_id"`
	TriggeredBy string           `gorm:"type:char(20);not null;default:null" json:"triggered_by"`
	TxUUID      string           `gorm:"type:char(64);" json:"tx_uuid,omitempty"`
	TxLen       int              `json:"tx_len,omitempty"`
	TxIndex     int              `json:"tx_index,omitempty"`
	Timestamp   uint64           `json:"timestamp"`
	Body        string           `json:"body,omitempty"`
}

// UnmarshalBody deserializes the event's JSON body to a given model
func (o *Event) UnmarshalBody(model interface{}) error {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Unmarshal([]byte(o.Body), model)
}

func (o *Event) TableName() string {
	return EventsTableName
}
