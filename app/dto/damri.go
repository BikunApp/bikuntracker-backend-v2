package dto

import "time"

type DamriAuthRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type DamriAuthResponseData struct {
	Id    string `json:"id"`
	Type  int    `json:"type"`
	Token string `json:"token"`
}

type DamriAuthResponse struct {
	RequestId string                `json:"request_id"`
	Code      int                   `json:"code"`
	Success   bool                  `json:"success"`
	Message   string                `json:"message"`
	Data      DamriAuthResponseData `json:"data"`
}

type BusCoordinate struct {
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

type DamriGetCoordinatesRequestBody struct {
	Imei []string `json:"imei"`
}

type DamriGetCoordinatesResponse struct {
	RequestId string          `json:"request_id"`
	Code      int             `json:"code"`
	Success   bool            `json:"success"`
	Message   string          `json:"message"`
	Data      []BusCoordinate `json:"data"`
}

type BusStatus struct {
	BusId       int       `json:"bus_id"`
	VehicleName string    `json:"vehicle_name"`
	Imei        string    `json:"imei"`
	IsActive    bool      `json:"is_active"`
	Color       string    `json:"color"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BikunAdminGetAllBusStatusResponse struct {
	Success bool        `json:"success"`
	Data    []BusStatus `json:"data"`
}
