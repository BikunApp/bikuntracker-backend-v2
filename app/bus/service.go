package bus

import (
	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
)

type service struct {
}

func NewService(repo interfaces.BusRepository) *service {
	return &service{}
}

// Add other services in the future here
