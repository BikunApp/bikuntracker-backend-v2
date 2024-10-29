package interfaces

import (
	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
)

type RMService interface {
	DetectLane(data map[string][]*models.BusCoordinate) (res dto.DetectRouteResponse, err error)
}
