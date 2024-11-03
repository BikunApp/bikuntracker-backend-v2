package bus

import (
	"context"

	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
	"github.com/FreeJ1nG/bikuntracker-backend/utils"
)

type seeder struct {
	repo interfaces.BusRepository
}

func NewSeeder(repo interfaces.BusRepository) *seeder {
	return &seeder{
		repo: repo,
	}
}

func (s *seeder) SeedBuses(ctx context.Context) {
	buses := utils.ReadJsonFromFixture[[]models.Bus]("./fixtures/bus/bus.json")
	err := s.repo.InsertBuses(ctx, buses)
	if err != nil {
		panic(err)
	}
}
