package bus

import (
	"log"
	"math"
	"time"

	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
	"github.com/gammazero/deque"
)

const (
	DQ_SIZE                        = 50
	DEFAULT_INTERPOLATION_INTERVAL = 1000 // 1 second in milliseconds
)

// InterpolationData holds data needed for coordinate interpolation
type InterpolationData struct {
	LastCoordinate     models.BusCoordinate
	NextCoordinate     *models.BusCoordinate // predicted next position
	Speed              float64               // km/h
	Bearing            float64               // degrees
	LastUpdate         time.Time
	IntervalSec        int     // seconds between real updates
	RouteColor         string  // current route color
	SegmentIndex       int     // current route segment index
	DistanceAlongRoute float64 // distance along current segment
}

type dqStore struct {
	dq      *deque.Deque[*models.BusCoordinate]
	counter int
}

type container struct {
	config              *models.Config
	rmService           interfaces.RMService
	damriService        interfaces.DamriService
	busService          interfaces.BusService
	busCoordinates      map[string]*models.BusCoordinate
	interpolationData   map[string]*InterpolationData // IMEI -> interpolation data
	storedBuses         map[string]*dqStore
	previousHalte       map[string]string // imei -> previous halte name
	activeLaps          map[string]bool   // imei -> whether bus has active lap
	currentPlates       map[string]string // imei -> current plate number
	interpolationTicker *time.Ticker
	stopInterpolation   chan bool
	routeMatcher        *RouteMatcher // for route-constrained interpolation
}

func NewContainer(
	config *models.Config,
	rmService interfaces.RMService,
	damriService interfaces.DamriService,
	busService interfaces.BusService,
) *container {
	// Initialize route matcher
	routeMatcher, err := NewRouteMatcher("routes-line-fixtures.txt")
	if err != nil {
		log.Printf("Warning: Failed to initialize route matcher: %v. Using fallback interpolation.", err)
		routeMatcher = nil
	} else {
		log.Printf("Route matcher initialized successfully with %d routes", len(routeMatcher.Routes))
	}

	return &container{
		config:            config,
		rmService:         rmService,
		damriService:      damriService,
		busService:        busService,
		busCoordinates:    make(map[string]*models.BusCoordinate),
		interpolationData: make(map[string]*InterpolationData),
		storedBuses:       make(map[string]*dqStore),
		previousHalte:     make(map[string]string),
		activeLaps:        make(map[string]bool),
		currentPlates:     make(map[string]string),
		stopInterpolation: make(chan bool),
		routeMatcher:      routeMatcher,
	}
}

// calculateDistance calculates the distance between two coordinates in meters using Haversine formula
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth's radius in meters

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLatRad := (lat2 - lat1) * math.Pi / 180
	deltaLonRad := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLatRad/2)*math.Sin(deltaLatRad/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLonRad/2)*math.Sin(deltaLonRad/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// calculateBearing calculates the bearing between two coordinates in degrees
func calculateBearing(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLonRad := (lon2 - lon1) * math.Pi / 180

	y := math.Sin(deltaLonRad) * math.Cos(lat2Rad)
	x := math.Cos(lat1Rad)*math.Sin(lat2Rad) - math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(deltaLonRad)

	bearing := math.Atan2(y, x) * 180 / math.Pi
	return math.Mod(bearing+360, 360)
}

// interpolateCoordinate calculates a new coordinate based on current position, bearing, and distance
func interpolateCoordinate(lat, lon, bearing, distanceMeters float64) (newLat, newLon float64) {
	const R = 6371000 // Earth's radius in meters

	latRad := lat * math.Pi / 180
	lonRad := lon * math.Pi / 180
	bearingRad := bearing * math.Pi / 180

	newLatRad := math.Asin(math.Sin(latRad)*math.Cos(distanceMeters/R) +
		math.Cos(latRad)*math.Sin(distanceMeters/R)*math.Cos(bearingRad))

	newLonRad := lonRad + math.Atan2(math.Sin(bearingRad)*math.Sin(distanceMeters/R)*math.Cos(latRad),
		math.Cos(distanceMeters/R)-math.Sin(latRad)*math.Sin(newLatRad))

	newLat = newLatRad * 180 / math.Pi
	newLon = newLonRad * 180 / math.Pi
	return
}

// startInterpolation starts the coordinate interpolation ticker
func (c *container) startInterpolation() {
	if c.interpolationTicker != nil {
		return // Already started
	}

	// Determine interpolation interval (hardcoded to 1000ms if not configured)
	intervalMs := c.config.InterpolationIntervalMs
	if intervalMs <= 0 {
		intervalMs = DEFAULT_INTERPOLATION_INTERVAL
	}
	interval := time.Duration(intervalMs) * time.Millisecond

	log.Printf("Starting coordinate interpolation system with %d ms intervals (hardcoded enabled)", intervalMs)

	c.interpolationTicker = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-c.interpolationTicker.C:
				c.updateInterpolatedCoordinates()
			case <-c.stopInterpolation:
				c.interpolationTicker.Stop()
				log.Printf("Coordinate interpolation system stopped")
				return
			}
		}
	}()
}

// stopInterpolationTicker stops the interpolation ticker
func (c *container) stopInterpolationTicker() {
	if c.interpolationTicker != nil {
		c.stopInterpolation <- true
		c.interpolationTicker = nil
	}
}

// CleanupInterpolation should be called when shutting down to clean up resources
func (c *container) CleanupInterpolation() {
	c.stopInterpolationTicker()
}

// updateInterpolatedCoordinates updates all bus coordinates with interpolated positions
func (c *container) updateInterpolatedCoordinates() {
	now := time.Now()

	for imei, interpData := range c.interpolationData {
		if coord, exists := c.busCoordinates[imei]; exists && interpData != nil {
			// Calculate time elapsed since last real update
			elapsed := now.Sub(interpData.LastUpdate)

			// Only interpolate if we have recent data (within last 30 seconds)
			if elapsed > 30*time.Second {
				continue
			}

			// Calculate how far we should have moved based on speed
			elapsedSeconds := elapsed.Seconds()
			speedMs := interpData.Speed * 1000 / 3600 // Convert km/h to m/s
			distanceToMove := speedMs * elapsedSeconds

			// Don't move too far (safety check)
			if distanceToMove > 200 { // Max 200m interpolation
				continue
			}

			// Use route-constrained interpolation if available
			if c.routeMatcher != nil && interpData.RouteColor != "" {
				newLat, newLon := c.interpolateAlongRoute(interpData, distanceToMove)
				if newLat != 0 && newLon != 0 {
					coord.Latitude = newLat
					coord.Longitude = newLon
					coord.GpsTime = now
					continue
				}
			}

			// Fallback to simple linear interpolation
			newLat, newLon := interpolateCoordinate(
				interpData.LastCoordinate.Latitude,
				interpData.LastCoordinate.Longitude,
				interpData.Bearing,
				distanceToMove,
			)

			// Update the coordinate with interpolated position
			coord.Latitude = newLat
			coord.Longitude = newLon
			coord.GpsTime = now
		}
	}
}

// interpolateAlongRoute interpolates position along the predefined route
func (c *container) interpolateAlongRoute(interpData *InterpolationData, distanceToMove float64) (lat, lon float64) {
	if c.routeMatcher == nil {
		return 0, 0
	}

	// Find the appropriate route based on current color
	route := c.routeMatcher.FindBestRoute(
		interpData.LastCoordinate.Latitude,
		interpData.LastCoordinate.Longitude,
		interpData.RouteColor,
	)

	if route == nil {
		return 0, 0
	}

	// Get the new position along the route
	newPoint := route.GetPointAtDistance(
		interpData.SegmentIndex,
		interpData.DistanceAlongRoute,
		distanceToMove,
	)

	return newPoint.Latitude, newPoint.Longitude
}

// updateInterpolationData updates interpolation data when new GPS coordinates are received
func (c *container) updateInterpolationData(coordinates map[string]*models.BusCoordinate) {
	now := time.Now()

	for imei, newCoord := range coordinates {
		if newCoord == nil {
			continue
		}

		// Get or create interpolation data for this bus
		if _, exists := c.interpolationData[imei]; !exists {
			c.interpolationData[imei] = &InterpolationData{}
		}

		interpData := c.interpolationData[imei]

		// Update route color
		interpData.RouteColor = newCoord.Color

		// If we have previous coordinate, calculate speed and bearing
		if interpData.LastUpdate.IsZero() {
			// First coordinate for this bus
			interpData.LastCoordinate = *newCoord
			interpData.LastUpdate = now
			interpData.Speed = float64(newCoord.Speed) // Use GPS speed if available
			interpData.Bearing = 0                     // Will be calculated on next update

			// Initialize route position if route matcher is available
			if c.routeMatcher != nil && newCoord.Color != "" && newCoord.Color != "grey" {
				route := c.routeMatcher.FindBestRoute(newCoord.Latitude, newCoord.Longitude, newCoord.Color)
				if route != nil {
					_, segmentIndex, distanceAlong, _ := route.FindClosestRoutePoint(newCoord.Latitude, newCoord.Longitude)
					interpData.SegmentIndex = segmentIndex
					interpData.DistanceAlongRoute = distanceAlong
				}
			}
		} else {
			// Calculate distance and time between coordinates
			distance := calculateDistance(
				interpData.LastCoordinate.Latitude,
				interpData.LastCoordinate.Longitude,
				newCoord.Latitude,
				newCoord.Longitude,
			)

			timeDiff := now.Sub(interpData.LastUpdate).Seconds()

			// Calculate speed in km/h
			if timeDiff > 0 {
				speedKmh := (distance / 1000) / (timeDiff / 3600)
				interpData.Speed = speedKmh
			}

			// Calculate bearing for direction
			interpData.Bearing = calculateBearing(
				interpData.LastCoordinate.Latitude,
				interpData.LastCoordinate.Longitude,
				newCoord.Latitude,
				newCoord.Longitude,
			)

			// Update route position if route matcher is available
			if c.routeMatcher != nil && newCoord.Color != "" && newCoord.Color != "grey" {
				route := c.routeMatcher.FindBestRoute(newCoord.Latitude, newCoord.Longitude, newCoord.Color)
				if route != nil {
					_, segmentIndex, distanceAlong, _ := route.FindClosestRoutePoint(newCoord.Latitude, newCoord.Longitude)
					interpData.SegmentIndex = segmentIndex
					interpData.DistanceAlongRoute = distanceAlong
				}
			}

			// Update with new coordinate
			interpData.LastCoordinate = *newCoord
			interpData.LastUpdate = now
		}
	}

	// Start interpolation if not already running
	c.startInterpolation()
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
