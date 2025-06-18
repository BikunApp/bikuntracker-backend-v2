package damri

import (
	"fmt"
	"time"

	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
)

const (
	MORNING_ROUTE   = 0
	NORMAL_ROUTE    = 1
	NOT_OPERATIONAL = 2
)

type service struct {
	config *models.Config
	util   interfaces.DamriUtil
}

func NewService(config *models.Config, util interfaces.DamriUtil) *service {
	return &service{
		config: config,
		util:   util,
	}
}

// Authenticate provides a dummy implementation for interface compatibility (not used with ws)
func (s *service) Authenticate() (token string, err error) {
	return "", nil
}

// GetOperationalStatus returns the operational status based on the latest bus data timestamp from the WebSocket.
func (s *service) GetOperationalStatus(busCoordinates map[string]*models.BusCoordinate) (int, error) {
	if len(busCoordinates) == 0 {
		return NOT_OPERATIONAL, nil
	}

	// Find the most recent GpsTime or fallback to requestedDate if available
	var latestTime time.Time
	for _, coord := range busCoordinates {
		if !coord.GpsTime.IsZero() && coord.GpsTime.After(latestTime) {
			latestTime = coord.GpsTime
		}
	}
	if latestTime.IsZero() {
		// If GpsTime is not set, fallback to current time
		latestTime = time.Now()
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		err = fmt.Errorf("unable to load Asia/Jakarta location")
		return NOT_OPERATIONAL, err
	}

	now := latestTime.In(loc)
	dayOfWeek := int(now.Weekday())
	currentTime := now.Hour()*60 + now.Minute()

	// If day is Monday - Friday
	if dayOfWeek >= 1 && dayOfWeek <= 5 {
		if currentTime >= s.util.GetHMInMinutes(6, 50) && currentTime < s.util.GetHMInMinutes(9, 0) {
			return MORNING_ROUTE, nil
		} else if currentTime >= s.util.GetHMInMinutes(9, 0) && currentTime < s.util.GetHMInMinutes(21, 30) {
			return NORMAL_ROUTE, nil
		} else {
			return NOT_OPERATIONAL, nil
		}
	}

	// If day is Saturday
	if dayOfWeek == 6 {
		if currentTime >= s.util.GetHMInMinutes(6, 50) && currentTime < s.util.GetHMInMinutes(16, 10) {
			return NORMAL_ROUTE, nil
		} else {
			return NOT_OPERATIONAL, nil
		}
	}

	// Sunday means that Bikun is not operational
	return NOT_OPERATIONAL, nil
}

func (s *service) GetBusCoordinates(imeiList []string) (res map[string]*models.BusCoordinate, err error) {
	// Dummy implementation for interface compatibility (not used with ws)
	return nil, nil
}
