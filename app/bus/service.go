package bus

import (
	"context"

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
