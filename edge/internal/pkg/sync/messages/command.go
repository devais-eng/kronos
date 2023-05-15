package messages

import "devais.it/kronos/internal/pkg/types"

type CommandType string

const (
	CommandGetVersion     CommandType = "GET_VERSION"
	CommandGetAllVersions CommandType = "GET_ALL_VERSIONS"
	CommandGetEntity      CommandType = "GET_ENTITY"
	CommandGetTelemetry   CommandType = "GET_TELEMETRY"
)

type ServerCommand struct {
	UUID        string                 `json:"uuid"`
	CommandType CommandType            `json:"command_type"`
	EntityType  types.EntityType       `json:"entity_type"`
	EntityID    string                 `json:"entity_id"`
	Body        map[string]interface{} `json:"body"`
}

type CommandResponse struct {
	UUID    string                 `json:"uuid"`
	Success bool                   `json:"success"`
	Error   string                 `json:"error,omitempty"`
	Body    map[string]interface{} `json:"body,omitempty"`
}
