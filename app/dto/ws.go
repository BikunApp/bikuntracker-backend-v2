package dto

import "github.com/FreeJ1nG/bikuntracker-backend/app/models"

type CoordinateBroadcastMessage struct {
	Coordinates       []models.BusCoordinate `json:"coordinates"`
	OperationalStatus int                    `json:"operationalStatus"`
}
