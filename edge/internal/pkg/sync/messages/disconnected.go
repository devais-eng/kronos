package messages

type Disconnected struct {
	DeviceID  string  `json:"device_id"`
	Timestamp *uint64 `json:"timestamp"`
}
