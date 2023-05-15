package messages

import "devais.it/kronos/internal/pkg/telemetry"

type Connected struct {
	DeviceID  string          `json:"device_id"`
	Timestamp *uint64         `json:"timestamp"`
	Telemetry *telemetry.Data `json:"telemetry"`
}
