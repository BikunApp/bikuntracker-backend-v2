package interfaces

import (
	"context"

	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
)

type BusContainer interface {
	RunCron() (err error)
}

type BusService interface {
}

type BusRepository interface {
	GetBuses(ctx context.Context) (res []models.Bus, err error)
	CreateBus(ctx context.Context, data dto.CreateBusRequestBody) (res *models.Bus, err error)
	UpdateBus(ctx context.Context, id string, data dto.UpdateBusRequestBody) (res *models.Bus, err error)
}
