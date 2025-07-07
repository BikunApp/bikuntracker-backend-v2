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
	UpdateCurrentHalteByImei(ctx context.Context, imei string, newHalte string) (*models.Bus, error)
	GetAllBuses(ctx context.Context) ([]models.Bus, error)
	// Lap history methods
	StartLap(ctx context.Context, imei string, routeColor string) (*models.BusLapHistory, error)
	EndLap(ctx context.Context, imei string) (*models.BusLapHistory, error)
	GetActiveLap(ctx context.Context, imei string) (*models.BusLapHistory, error)
	GetFilteredLapHistory(ctx context.Context, filter dto.LapHistoryFilter) ([]models.BusLapHistory, error)
	GetFilteredLapHistoryCount(ctx context.Context, filter dto.LapHistoryFilter) (int, error)
}

type BusRepository interface {
	GetBuses(ctx context.Context) (res []models.Bus, err error)
	CreateBus(ctx context.Context, data dto.CreateBusRequestBody) (res *models.Bus, err error)
	UpdateBus(ctx context.Context, whereData *models.WhereData, data dto.UpdateBusRequestBody) (res *models.Bus, err error)
	DeleteBus(ctx context.Context, id string) (err error)
	InsertBuses(ctx context.Context, data []models.Bus) (err error)
	// Lap history methods
	CreateLapHistory(ctx context.Context, lapHistory *models.BusLapHistory) (*models.BusLapHistory, error)
	UpdateLapHistory(ctx context.Context, id int, endTime interface{}) (*models.BusLapHistory, error)
	GetActiveLapByImei(ctx context.Context, imei string) (*models.BusLapHistory, error)
	GetLapHistoryByImei(ctx context.Context, imei string) ([]models.BusLapHistory, error)
	GetFilteredLapHistory(ctx context.Context, filter dto.LapHistoryFilter) ([]models.BusLapHistory, error)
	GetFilteredLapHistoryCount(ctx context.Context, filter dto.LapHistoryFilter) (int, error)
	// Debug methods
	GetLapHistoryCount(ctx context.Context) (int, error)
}
