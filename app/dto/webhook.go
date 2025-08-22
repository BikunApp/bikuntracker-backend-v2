package dto

// WebhookPayload represents the incoming webhook structure for location updates
type WebhookPayload struct {
	EventTime int64  `json:"event_time"`
	EventID   int    `json:"event_id"`
	EventName string `json:"event_name"`
	EventType string `json:"event_type"`
	Event     struct {
		Data WebhookData `json:"data"`
	} `json:"event"`
}

// WebhookData contains the actual location data
type WebhookData struct {
	LicensePlate string  `json:"license_plate"`
	HullNo       string  `json:"hull_no"`
	IMEI         string  `json:"imei"`
	Driver       string  `json:"driver"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Speed        float64 `json:"speed"`
	Direction    float64 `json:"direction"`
	EngineOn     bool    `json:"engine_on"`
	LastPacket   string  `json:"last_packet"`
	LastReceive  string  `json:"last_receive"`
	LastMotion   string  `json:"last_motion"`
	MotionStatus string  `json:"motion_status"`
}
