package models

import "time"

type BusLapHistory struct {
	ID         int        `json:"id"`
	BusID      int        `json:"bus_id"`
	IMEI       string     `json:"imei"`
	LapNumber  int        `json:"lap_number"`
	StartTime  time.Time  `json:"start_time"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	RouteColor string     `json:"route_color"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
