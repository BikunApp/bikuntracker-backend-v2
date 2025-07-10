package bus

import (
	"context"
	"errors"
	"time"

	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
)

type service struct {
	repo interfaces.BusRepository
}

func NewService(repo interfaces.BusRepository) *service {
	return &service{
		repo: repo,
	}
}

func (s *service) UpdateBusColorByImei(ctx context.Context, imei string, newColor string) (*models.Bus, error) {
	return s.repo.UpdateBus(
		ctx,
		&models.WhereData{
			FieldName: "imei",
			Value:     imei,
		},
		dto.UpdateBusRequestBody{
			Color: &newColor,
		},
	)
}

func (s *service) UpdateCurrentHalteByImei(ctx context.Context, imei string, newHalte string) (*models.Bus, error) {
	return s.repo.UpdateBus(
		ctx,
		&models.WhereData{
			FieldName: "imei",
			Value:     imei,
		},
		dto.UpdateBusRequestBody{
			CurrentHalte: &newHalte,
		},
	)
}

func (s *service) GetAllBuses(ctx context.Context) ([]models.Bus, error) {
	return s.repo.GetBuses(ctx)
}

// Lap tracking methods
func (s *service) StartLap(ctx context.Context, imei string, routeColor string) (*models.BusLapHistory, error) {
	// Get bus info to get bus_id
	buses, err := s.repo.GetBuses(ctx)
	if err != nil {
		return nil, err
	}

	var busID int
	for _, bus := range buses {
		if bus.Imei == imei {
			busID = bus.Id
			break
		}
	}

	if busID == 0 {
		return nil, errors.New("no bus found with the given IMEI")
	}

	// Get current lap number by checking existing laps
	existingLaps, _ := s.repo.GetLapHistoryByImei(ctx, imei)
	lapNumber := len(existingLaps) + 1

	lapHistory := &models.BusLapHistory{
		BusID:      busID,
		IMEI:       imei,
		LapNumber:  lapNumber,
		StartTime:  time.Now(),
		RouteColor: routeColor,
	}

	return s.repo.CreateLapHistory(ctx, lapHistory)
}

func (s *service) EndLap(ctx context.Context, imei string) (*models.BusLapHistory, error) {
	activeLap, err := s.repo.GetActiveLapByImei(ctx, imei)
	if err != nil {
		return nil, err
	}

	if activeLap == nil {
		return nil, nil // No active lap to end
	}

	endTime := time.Now()
	return s.repo.UpdateLapHistory(ctx, activeLap.ID, endTime)
}

func (s *service) GetActiveLap(ctx context.Context, imei string) (*models.BusLapHistory, error) {
	return s.repo.GetActiveLapByImei(ctx, imei)
}

func (s *service) GetFilteredLapHistory(ctx context.Context, filter dto.LapHistoryFilter) ([]models.BusLapHistory, error) {
	return s.repo.GetFilteredLapHistory(ctx, filter)
}

func (s *service) GetFilteredLapHistoryCount(ctx context.Context, filter dto.LapHistoryFilter) (int, error) {
	return s.repo.GetFilteredLapHistoryCount(ctx, filter)
}
