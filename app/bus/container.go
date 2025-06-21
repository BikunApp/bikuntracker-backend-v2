package bus

import (
	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
	"github.com/gammazero/deque"
)

const (
	DQ_SIZE = 50
)

type dqStore struct {
	dq      *deque.Deque[*models.BusCoordinate]
	counter int
}

type container struct {
	config         *models.Config
	rmService      interfaces.RMService
	damriService   interfaces.DamriService
	busService     interfaces.BusService
	busCoordinates map[string]*models.BusCoordinate
	storedBuses    map[string]*dqStore
	previousHalte  map[string]string // imei -> previous halte name
}

func NewContainer(
	config *models.Config,
	rmService interfaces.RMService,
	damriService interfaces.DamriService,
	busService interfaces.BusService,
) *container {
	return &container{
		config:         config,
		rmService:      rmService,
		damriService:   damriService,
		busService:     busService,
		busCoordinates: make(map[string]*models.BusCoordinate),
		storedBuses:    make(map[string]*dqStore),
		previousHalte:  make(map[string]string),
	}
}

func (c *container) GetBusCoordinates() (res []models.BusCoordinate) {
	res = make([]models.BusCoordinate, 0)
	for _, busCoordinate := range c.busCoordinates {
		res = append(res, *busCoordinate)
	}
	return
}

func (c *container) GetBusCoordinatesMap() map[string]*models.BusCoordinate {
	return c.busCoordinates
}
