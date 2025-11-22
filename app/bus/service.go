package bus

import (
	"context"
	"errors"
	"strings"
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

// Helper function to convert lap history time fields to UTC
func (s *service) convertLapHistoryToUTC(lap *models.BusLapHistory) {
	if lap == nil {
		return
	}
	lap.StartTime = lap.StartTime.UTC()
	if lap.EndTime != nil {
		utcEndTime := lap.EndTime.UTC()
		lap.EndTime = &utcEndTime
	}
	lap.CreatedAt = lap.CreatedAt.UTC()
	lap.UpdatedAt = lap.UpdatedAt.UTC()
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

func (s *service) UpdateBusPlateNumberByImei(ctx context.Context, imei string, plateNumber string) (*models.Bus, error) {
	return s.repo.UpdateBus(
		ctx,
		&models.WhereData{
			FieldName: "imei",
			Value:     imei,
		},
		dto.UpdateBusRequestBody{
			PlateNumber: &plateNumber,
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

	// Initialize halte visit history with starting point and timestamp
	jakarta, _ := time.LoadLocation("Asia/Jakarta")
	startTimeJakarta := time.Now().In(jakarta)
	initialHalteHistory := "Asrama UI [" + startTimeJakarta.Format("2006-01-02 15:04:05") + "]"

	lapHistory := &models.BusLapHistory{
		BusID:             busID,
		IMEI:              imei,
		LapNumber:         lapNumber,
		StartTime:         time.Now(),
		RouteColor:        routeColor,
		HalteVisitHistory: initialHalteHistory,
	}

	result, err := s.repo.CreateLapHistory(ctx, lapHistory)
	if err != nil {
		return nil, err
	}

	// Convert all time fields to UTC (Zulu time)
	s.convertLapHistoryToUTC(result)

	return result, nil
}

func (s *service) EndLap(ctx context.Context, imei string) (*models.BusLapHistory, error) {
	activeLap, err := s.repo.GetActiveLapByImei(ctx, imei)
	if err != nil {
		return nil, err
	}

	if activeLap == nil {
		return nil, nil // No active lap to end
	}

	// Get the current bus to fetch its latest color
	buses, err := s.repo.GetBuses(ctx)
	if err != nil {
		return nil, err
	}

	var currentBusColor string = "grey" // Default fallback color
	for _, bus := range buses {
		if bus.Imei == imei {
			currentBusColor = bus.Color
			break
		}
	}

	endTime := time.Now()
	result, err := s.repo.UpdateLapHistoryWithColor(ctx, activeLap.ID, endTime, currentBusColor)
	if err != nil {
		return nil, err
	}

	// Convert all time fields to UTC (Zulu time)
	s.convertLapHistoryToUTC(result)

	return result, nil
}

func (s *service) GetActiveLap(ctx context.Context, imei string) (*models.BusLapHistory, error) {
	lap, err := s.repo.GetActiveLapByImei(ctx, imei)
	if err != nil {
		return nil, err
	}

	// Convert all time fields to UTC (Zulu time)
	s.convertLapHistoryToUTC(lap)

	return lap, nil
}

func (s *service) GetFilteredLapHistory(ctx context.Context, filter dto.LapHistoryFilter) ([]models.BusLapHistory, error) {
	laps, err := s.repo.GetFilteredLapHistory(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Convert all time fields to UTC (Zulu time)
	for i := range laps {
		s.convertLapHistoryToUTC(&laps[i])
	}

	return laps, nil
}

func (s *service) GetFilteredLapHistoryCount(ctx context.Context, filter dto.LapHistoryFilter) (int, error) {
	return s.repo.GetFilteredLapHistoryCount(ctx, filter)
}

func (s *service) AddHalteVisitToActiveLap(ctx context.Context, imei string, halteName string) error {
	// Get the active lap for this bus
	activeLap, err := s.repo.GetActiveLapByImei(ctx, imei)
	if err != nil {
		return err
	}

	if activeLap == nil {
		// No active lap, nothing to update
		return nil
	}

	// Build the new halte visit history with timestamp
	jakarta, _ := time.LoadLocation("Asia/Jakarta")
	currentTimeJakarta := time.Now().In(jakarta)
	halteWithTimestamp := halteName + " [" + currentTimeJakarta.Format("2006-01-02 15:04:05") + "]"

	var newHalteHistory string
	if activeLap.HalteVisitHistory == "" {
		newHalteHistory = halteWithTimestamp
	} else {
		// Check if this halte is already the last one to avoid duplicates
		existingHaltes := strings.Split(activeLap.HalteVisitHistory, " -> ")
		if len(existingHaltes) > 0 {
			// Extract just the halte name from the last entry (without timestamp)
			lastHalte := existingHaltes[len(existingHaltes)-1]
			if strings.Contains(lastHalte, " [") {
				lastHalte = strings.Split(lastHalte, " [")[0]
			}
			if lastHalte == halteName {
				// Same halte as last, don't add duplicate
				return nil
			}
		}
		newHalteHistory = activeLap.HalteVisitHistory + " -> " + halteWithTimestamp
	}

	// Update the halte visit history
	return s.repo.UpdateLapHistoryHalteVisits(ctx, activeLap.ID, newHalteHistory)
}
