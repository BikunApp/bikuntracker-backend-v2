package bus

import (
	"context"
	"log"

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
	currentPlates  map[string]string // imei -> current plate number
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
		currentPlates:  make(map[string]string),
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

// InitRuntimeState initializes runtime caches from database (colors, active laps, plates)
func (c *container) InitRuntimeState() {
	ctx := context.Background()
	buses, err := c.busService.GetAllBuses(ctx)
	if err != nil {
		log.Printf("Failed to initialize runtime state: %v", err)
		return
	}
	for _, b := range buses {
		_, _ = c.busService.UpdateBusColorByImei(ctx, b.Imei, "grey")
		activeLap, _ := c.busService.GetActiveLap(ctx, b.Imei)
		c.activeLaps[b.Imei] = activeLap != nil
		c.currentPlates[b.Imei] = b.PlateNumber
	}
}

// ApplyExternalCoordinates allows feeding new coordinates from an external source (webhook)
// without modifying business logic. It reuses the same pipeline as WS ingestion.
func (c *container) ApplyExternalCoordinates(coords map[string]*models.BusCoordinate) {
	// Update colors based on halte transitions
	c.updateBusColors(coords)
	// Store into rolling windows for lane detection
	c.insertFetchedData(coords)
	// Update halte visits and lap start/end
	c.updateHalteVisits(context.Background(), coords)
	// Possibly change bus lane via RM service
	if err := c.possiblyChangeBusLane(); err != nil {
		log.Printf("Unable to change bus lane: %s", err.Error())
	}
	// Optional logs
	c.logCsvIfNeeded(coords)
	// Replace runtime map
	c.busCoordinates = coords
}
