package interfaces

import "github.com/FreeJ1nG/bikuntracker-backend/app/dto"

type DamriService interface {
	Authenticate() (token string, err error)
	GetAllBusStatus() (res []dto.BusStatus, err error)
	GetBusCoordinates(imeiList []string) (res []dto.BusCoordinate, err error)
}

type DamriUtil interface {
	GetHMInMinutes(hours int, minutes int) int
}
