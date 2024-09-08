package interfaces

import "github.com/FreeJ1nG/bikuntracker-backend/app/models"

type DamriService interface {
	Authenticate() (token string, err error)
	GetAllBusStatus() (res []models.BusStatus, err error)
	GetBusCoordinates(imeiList []string) (res []models.BusCoordinate, err error)
}

type DamriUtil interface {
	GetHMInMinutes(hours int, minutes int) int
}
