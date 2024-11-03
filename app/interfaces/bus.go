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
	UpdateBusColorByImei(ctx context.Context, imei string, newColor string) (*models.Bus, error)
	GetAllBuses(ctx context.Context) ([]models.Bus, error)
}

type BusRepository interface {
	GetBuses(ctx context.Context) (res []models.Bus, err error)
	CreateBus(ctx context.Context, data dto.CreateBusRequestBody) (res *models.Bus, err error)
	UpdateBus(ctx context.Context, whereData *models.WhereData, data dto.UpdateBusRequestBody) (res *models.Bus, err error)
	DeleteBus(ctx context.Context, id string) (err error)
	InsertBuses(ctx context.Context, data []models.Bus) (err error)
}
