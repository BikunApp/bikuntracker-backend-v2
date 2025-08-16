package models

import "time"

type BusCoordinate struct {
	Id            int       `json:"id"`
	Color         string    `json:"color"`
	Imei          string    `json:"imei"`
	VehicleName   string    `json:"vehicle_name"`
	BusNumber     string    `json:"bus_number"`
	PlateNumber   string    `json:"plate_number"`
	Longitude     float64   `json:"longitude"`
	Latitude      float64   `json:"latitude"`
	Status        string    `json:"status"`
	Speed         int       `json:"speed"`
	TotalMileage  float64   `json:"total_mileage"`
	GpsTime       time.Time `json:"gps_time"`
	CurrentHalte  string    `json:"current_halte"`
	StatusMessage string    `json:"message"`
	NextHalte     string    `json:"next_halte"`
}

type Bus struct {
	Id           int    `json:"id"`
	VehicleNo    string `json:"vehicle_no"`
	Imei         string `json:"imei"`
	IsActive     bool   `json:"is_active"`
	Color        string `json:"color"`
	BusNumber    string `json:"bus_number"`
	PlateNumber  string `json:"plate_number"`
	CurrentHalte string `json:"current_halte"`
	NextHalte    string `json:"next_halte"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}
