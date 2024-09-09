package interfaces

import "github.com/FreeJ1nG/bikuntracker-backend/app/dto"

type DamriService interface {
	Authenticate() (response dto.DamriAuthResponse, err error)
	GetAllBusStatus() (res []dto.BusStatus, err error)
}
