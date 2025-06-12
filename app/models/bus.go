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
	CurrentHalte string    `json:"current_halte"`
	NextHalte    string    `json:"next_halte"`
}

type Bus struct {
	Id        int    `json:"id"`
	VehicleNo string `json:"vehicle_no"`
	Imei      string `json:"imei"`
	IsActive  bool   `json:"is_active"`
	Color     string `json:"color"`
	CreatedAt uint   `json:"created_at"`
	UpdatedAt uint   `json:"updated_at"`
}
