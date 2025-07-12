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
	activeLaps     map[string]bool   // imei -> whether bus has active lap
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
		activeLaps:     make(map[string]bool),
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

func (c *container) UpdateRuntimeBusColor(imei string, color string) error {
	if coord, exists := c.busCoordinates[imei]; exists {
		coord.Color = color
		return nil
	}
	// Return error if bus not found in runtime coordinates, but don't fail the request
	return nil
}

func (c *container) RunCron() (err error) {
	// Implementation would go here
	return nil
}
