package dto

import "github.com/FreeJ1nG/bikuntracker-backend/app/models"

type GetBusesResponse = []models.Bus

type CreateBusRequestBody struct {
	VehicleNo string `json:"vehicle_no"`
	Imei      string `json:"imei"`
	IsActive  bool   `json:"is_active"`
	Color     string `json:"color"`
}

type UpdateBusRequestBody struct {
	VehicleNo *string `json:"vehicle_no,omitempty"`
	Imei      *string `json:"imei,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
	Color     *string `json:"color,omitempty"`
}
