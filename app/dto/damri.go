package dto

import "github.com/FreeJ1nG/bikuntracker-backend/app/models"

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

type DamriGetCoordinatesRequestBody struct {
	Imei []string `json:"imei"`
}

type DamriGetCoordinatesResponse struct {
	RequestId string                 `json:"request_id"`
	Code      int                    `json:"code"`
	Success   bool                   `json:"success"`
	Message   string                 `json:"message"`
	Data      []models.BusCoordinate `json:"data"`
}

type BikunAdminGetAllBusStatusResponse struct {
	Success bool               `json:"success"`
	Data    []models.BusStatus `json:"data"`
}
