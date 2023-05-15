package types

type EventType string

const (
	EventEntityCreated EventType = "ENTITY_CREATED"
	EventEntityUpdated EventType = "ENTITY_UPDATED"
	EventEntityDeleted EventType = "ENTITY_DELETED"
)
