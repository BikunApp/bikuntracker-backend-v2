package models

import "time"

type BusCoordinate struct {
	Id           int       `json:"id"`
	Color        string    `json:"color"`
	Imei         string    `json:"imei"`
	VehicleName  string    `json:"vehicle_name"`
	Longitude    float64   `json:"longitude"`
	Latitude     float64   `json:"latitude"`
	Status       string    `json:"status"`
	Speed        int       `json:"speed"`
	TotalMileage float64   `json:"total_mileage"`
	GpsTime      time.Time `json:"gps_time"`
}

type BusStatus struct {
	BusId       int       `json:"bus_id"`
	VehicleName string    `json:"vehicle_name"`
	Imei        string    `json:"imei"`
	IsActive    bool      `json:"is_active"`
	Color       string    `json:"color"`
	UpdatedAt   time.Time `json:"updated_at"`
}
