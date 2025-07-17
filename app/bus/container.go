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
	LastRealCoordinate models.BusCoordinate  // Last actual GPS coordinate
	CurrentCoordinate  models.BusCoordinate  // Current position (real or interpolated)
	NextCoordinate     *models.BusCoordinate // predicted next position
	Speed              float64               // km/h (capped for realistic movement)
	Bearing            float64               // degrees
	LastRealUpdate     time.Time             // Last real GPS update time
	LastInterpolation  time.Time             // Last interpolation update time
	IntervalSec        int                   // seconds between real updates
	RouteColor         string                // current route color
	SegmentIndex       int                   // current route segment index
	DistanceAlongRoute float64               // distance along current segment
	TotalRouteDistance float64               // total distance along entire route
	IsInterpolating    bool                  // whether currently showing interpolated position
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
			// Calculate time elapsed since last real GPS update
			elapsed := now.Sub(interpData.LastRealUpdate)

			// Only interpolate if we have recent data (within last 12 seconds)
			if elapsed > 12*time.Second {
				continue
			}

			// Calculate time since last interpolation step
			interpolationElapsed := now.Sub(interpData.LastInterpolation)
			if interpolationElapsed < 300*time.Millisecond {
				continue // Update very frequently for ultra-smooth movement
			}

			// Use realistic bus speed (max 20 km/h in campus)
			maxSpeed := 20.0 // km/h (balanced for campus buses)
			effectiveSpeed := interpData.Speed
			if effectiveSpeed > maxSpeed {
				effectiveSpeed = maxSpeed
			}
			if effectiveSpeed < 0.0 {
				effectiveSpeed = 0.0 // Minimum realistic speed (buses can stop at haltes)
			}

			// Calculate distance to move in this interpolation step (not total elapsed)
			stepSeconds := interpolationElapsed.Seconds()
			speedMs := effectiveSpeed * 1000 / 3600 // Convert km/h to m/s
			distanceToMove := speedMs * stepSeconds

			// Limit movement per step to very small increments for smooth movement
			if distanceToMove > 3 { // Max 3m per interpolation step (very small steps)
				distanceToMove = 3
			}

			// Don't move if speed is too low (bus is essentially stopped)
			if effectiveSpeed < 1.0 { // If moving slower than 1 km/h, consider it stopped
				distanceToMove = 0
			}

			// Calculate distance to real GPS coordinate to avoid overshooting
			distanceToReal := calculateDistance(
				interpData.CurrentCoordinate.Latitude,
				interpData.CurrentCoordinate.Longitude,
				interpData.LastRealCoordinate.Latitude,
				interpData.LastRealCoordinate.Longitude,
			)

			// If we're very close to the real coordinate, use smaller steps
			if distanceToReal < 10 && distanceToMove > distanceToReal/3 {
				distanceToMove = distanceToReal / 3 // Move 1/3 of remaining distance
			}

			// Use route-constrained interpolation if available and route color is valid
			if c.routeMatcher != nil && interpData.RouteColor != "" && interpData.RouteColor != "grey" {
				newLat, newLon := c.interpolateAlongRoute(interpData, distanceToMove)
				if newLat != 0 && newLon != 0 {
					// Verify the new position is reasonable (not too far from route)
					route := c.routeMatcher.FindBestRoute(newLat, newLon, interpData.RouteColor)
					if route != nil {
						closestPoint, _, _, _ := route.FindClosestRoutePoint(newLat, newLon)
						distanceFromRoute := calculateDistance(newLat, newLon, closestPoint.Latitude, closestPoint.Longitude)
						// Only use route interpolation if we stay close to the route (within 30m)
						if distanceFromRoute < 30 {
							coord.Latitude = newLat
							coord.Longitude = newLon
							coord.GpsTime = now
							interpData.CurrentCoordinate.Latitude = newLat
							interpData.CurrentCoordinate.Longitude = newLon
							interpData.LastInterpolation = now
							interpData.IsInterpolating = true
							continue
						}
					}
				}
			}

			// Fallback to simple linear interpolation if route-constrained fails
			// This ensures buses keep moving even if route matching has issues
			if distanceToMove > 0 {
				// Calculate bearing towards the real GPS coordinate for more accurate movement
				bearingToReal := calculateBearing(
					interpData.CurrentCoordinate.Latitude,
					interpData.CurrentCoordinate.Longitude,
					interpData.LastRealCoordinate.Latitude,
					interpData.LastRealCoordinate.Longitude,
				)

				// Use bearing towards real coordinate if we're close, otherwise use calculated bearing
				bearingToUse := interpData.Bearing
				distanceToReal := calculateDistance(
					interpData.CurrentCoordinate.Latitude,
					interpData.CurrentCoordinate.Longitude,
					interpData.LastRealCoordinate.Latitude,
					interpData.LastRealCoordinate.Longitude,
				)

				if distanceToReal < 50 { // Within 50m, move towards real coordinate
					bearingToUse = bearingToReal
				}

				newLat, newLon := interpolateCoordinate(
					interpData.CurrentCoordinate.Latitude,
					interpData.CurrentCoordinate.Longitude,
					bearingToUse,
					distanceToMove,
				)

				// Update the coordinate with interpolated position
				coord.Latitude = newLat
				coord.Longitude = newLon
				coord.GpsTime = now
				interpData.CurrentCoordinate.Latitude = newLat
				interpData.CurrentCoordinate.Longitude = newLon
				interpData.LastInterpolation = now
				interpData.IsInterpolating = true
			}
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
		interpData.CurrentCoordinate.Latitude,
		interpData.CurrentCoordinate.Longitude,
		interpData.RouteColor,
	)

	if route == nil {
		return 0, 0
	}

	// First verify current position is actually on the route
	_, currentSegmentIndex, currentDistanceAlong, _ := route.FindClosestRoutePoint(
		interpData.CurrentCoordinate.Latitude,
		interpData.CurrentCoordinate.Longitude,
	)

	// Use the more accurate current position on route
	interpData.SegmentIndex = currentSegmentIndex
	interpData.DistanceAlongRoute = currentDistanceAlong

	// Get the new position along the route with very small movement
	newPoint := route.GetPointAtDistance(
		interpData.SegmentIndex,
		interpData.DistanceAlongRoute,
		distanceToMove,
	)

	// Verify new point is reasonable
	if newPoint.Latitude == 0 || newPoint.Longitude == 0 {
		return 0, 0
	}

	// Update route position for next interpolation
	_, newSegmentIndex, newDistanceAlong, newTotalDistance := route.FindClosestRoutePoint(newPoint.Latitude, newPoint.Longitude)
	interpData.SegmentIndex = newSegmentIndex
	interpData.DistanceAlongRoute = newDistanceAlong
	interpData.TotalRouteDistance = newTotalDistance

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
		if interpData.LastRealUpdate.IsZero() {
			// First coordinate for this bus
			interpData.LastRealCoordinate = *newCoord
			interpData.CurrentCoordinate = *newCoord
			interpData.LastRealUpdate = now
			interpData.LastInterpolation = now
			interpData.Speed = float64(newCoord.Speed) // Use GPS speed if available
			interpData.Bearing = 0                     // Will be calculated on next update
			interpData.IsInterpolating = false

			// Make sure the actual bus coordinate exists and is initialized
			if coord, exists := c.busCoordinates[imei]; exists {
				coord.Latitude = newCoord.Latitude
				coord.Longitude = newCoord.Longitude
				coord.GpsTime = newCoord.GpsTime
			}

			// Initialize route position if route matcher is available
			if c.routeMatcher != nil && newCoord.Color != "" && newCoord.Color != "grey" {
				route := c.routeMatcher.FindBestRoute(newCoord.Latitude, newCoord.Longitude, newCoord.Color)
				if route != nil {
					_, segmentIndex, distanceAlong, totalDistance := route.FindClosestRoutePoint(newCoord.Latitude, newCoord.Longitude)
					interpData.SegmentIndex = segmentIndex
					interpData.DistanceAlongRoute = distanceAlong
					interpData.TotalRouteDistance = totalDistance
				}
			}
		} else {
			// Calculate distance and time between real GPS coordinates
			distance := calculateDistance(
				interpData.LastRealCoordinate.Latitude,
				interpData.LastRealCoordinate.Longitude,
				newCoord.Latitude,
				newCoord.Longitude,
			)

			timeDiff := now.Sub(interpData.LastRealUpdate).Seconds()

			// Calculate realistic speed in km/h
			if timeDiff > 0 {
				speedKmh := (distance / 1000) / (timeDiff / 3600)
				// Cap speed to realistic bus speeds (max 50 km/h)
				if speedKmh > 30 {
					speedKmh = 30
				}
				interpData.Speed = speedKmh
			}

			// Calculate bearing for direction
			interpData.Bearing = calculateBearing(
				interpData.LastRealCoordinate.Latitude,
				interpData.LastRealCoordinate.Longitude,
				newCoord.Latitude,
				newCoord.Longitude,
			)

			// Update route position if route matcher is available
			if c.routeMatcher != nil && newCoord.Color != "" && newCoord.Color != "grey" {
				route := c.routeMatcher.FindBestRoute(newCoord.Latitude, newCoord.Longitude, newCoord.Color)
				if route != nil {
					_, segmentIndex, distanceAlong, totalDistance := route.FindClosestRoutePoint(newCoord.Latitude, newCoord.Longitude)
					interpData.SegmentIndex = segmentIndex
					interpData.DistanceAlongRoute = distanceAlong
					interpData.TotalRouteDistance = totalDistance
				}
			}

			// Smooth transition: if we were interpolating, don't jump directly to new GPS position
			// Instead, update the targets and let interpolation catch up
			currentDistance := calculateDistance(
				interpData.CurrentCoordinate.Latitude,
				interpData.CurrentCoordinate.Longitude,
				newCoord.Latitude,
				newCoord.Longitude,
			)

			// If the new GPS position is very close to our current interpolated position,
			// or if we haven't been interpolating, update directly
			if currentDistance < 30 || !interpData.IsInterpolating {
				interpData.CurrentCoordinate = *newCoord
				// Also update the actual bus coordinate that gets returned to frontend
				if coord, exists := c.busCoordinates[imei]; exists {
					coord.Latitude = newCoord.Latitude
					coord.Longitude = newCoord.Longitude
					coord.GpsTime = newCoord.GpsTime
				}
			}
			// Otherwise, let interpolation gradually move towards the new real position

			// Update with new real coordinate
			interpData.LastRealCoordinate = *newCoord
			interpData.LastRealUpdate = now
			interpData.IsInterpolating = false // Reset interpolation flag
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
