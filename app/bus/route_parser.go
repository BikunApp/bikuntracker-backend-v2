package bus

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// RoutePoint represents a coordinate point on a route
type RoutePoint struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// Route represents a complete bus route with all its coordinate points
type Route struct {
	Name   string       `json:"name"`
	Points []RoutePoint `json:"points"`
}

// RouteSegment represents a segment between two consecutive route points
type RouteSegment struct {
	Start    RoutePoint
	End      RoutePoint
	Distance float64 // distance in meters
	Bearing  float64 // bearing in degrees
}

// RouteMatcher helps find the closest route and position for interpolation
type RouteMatcher struct {
	Routes map[string]*Route // route color -> route data
}

// NewRouteMatcher creates a new route matcher by parsing the routes file
func NewRouteMatcher(routesFilePath string) (*RouteMatcher, error) {
	routes := make(map[string]*Route)

	// Parse the routes file
	blueNormal, err := parseRouteFromFile(routesFilePath, "BLUE_NORMAL_ROUTE")
	if err != nil {
		return nil, fmt.Errorf("failed to parse blue normal route: %v", err)
	}
	routes["blue"] = blueNormal

	blueMorning, err := parseRouteFromFile(routesFilePath, "BLUE_MORNING_ROUTE")
	if err != nil {
		return nil, fmt.Errorf("failed to parse blue morning route: %v", err)
	}
	routes["express-blue"] = blueMorning

	redNormal, err := parseRouteFromFile(routesFilePath, "RED_NORMAL_ROUTE")
	if err != nil {
		return nil, fmt.Errorf("failed to parse red normal route: %v", err)
	}
	routes["red"] = redNormal

	redMorning, err := parseRouteFromFile(routesFilePath, "RED_MORNING_ROUTE")
	if err != nil {
		return nil, fmt.Errorf("failed to parse red morning route: %v", err)
	}
	routes["express-red"] = redMorning

	return &RouteMatcher{Routes: routes}, nil
}

// parseRouteFromFile extracts route coordinates from the JavaScript/TypeScript fixture file
func parseRouteFromFile(filePath, routeName string) (*Route, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var inRoute bool
	var inCoordinates bool
	var coordinates []RoutePoint
	var braceCount int

	// Regex to match coordinate pairs: [longitude, latitude]
	coordRegex := regexp.MustCompile(`\[\s*([-\d.]+)\s*,\s*([-\d.]+)\s*\]`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Check if we're starting the target route
		if strings.Contains(line, routeName) {
			inRoute = true
			continue
		}

		if !inRoute {
			continue
		}

		// Track brace nesting to know when we exit the route
		braceCount += strings.Count(line, "{") - strings.Count(line, "}")

		// Look for coordinates array
		if strings.Contains(line, "coordinates:") {
			inCoordinates = true
			continue
		}

		if inCoordinates {
			// Extract coordinate pairs from this line
			matches := coordRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) == 3 {
					lng, err1 := strconv.ParseFloat(match[1], 64)
					lat, err2 := strconv.ParseFloat(match[2], 64)
					if err1 == nil && err2 == nil {
						coordinates = append(coordinates, RoutePoint{
							Longitude: lng,
							Latitude:  lat,
						})
					}
				}
			}

			// Check if we've finished reading coordinates
			if strings.Contains(line, "],") && strings.Contains(line, "type:") {
				inCoordinates = false
			}
		}

		// Exit route when braces are balanced (route definition complete)
		if inRoute && braceCount <= 0 && len(coordinates) > 0 {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(coordinates) == 0 {
		return nil, fmt.Errorf("no coordinates found for route %s", routeName)
	}

	return &Route{
		Name:   routeName,
		Points: coordinates,
	}, nil
}

// FindClosestRoutePoint finds the closest point on a route to the given coordinate
func (r *Route) FindClosestRoutePoint(lat, lng float64) (closestPoint RoutePoint, segmentIndex int, distanceAlongSegment float64, totalDistance float64) {
	minDistance := math.Inf(1)
	bestPoint := RoutePoint{}
	bestSegmentIndex := 0
	bestDistanceAlong := 0.0

	for i := 0; i < len(r.Points)-1; i++ {
		start := r.Points[i]
		end := r.Points[i+1]

		// Find closest point on this segment
		closest, distAlong := closestPointOnSegment(lat, lng, start, end)
		distance := calculateDistance(lat, lng, closest.Latitude, closest.Longitude)

		if distance < minDistance {
			minDistance = distance
			bestPoint = closest
			bestSegmentIndex = i
			bestDistanceAlong = distAlong
		}
	}

	// Calculate total distance: sum of all complete segments before bestSegmentIndex + bestDistanceAlong
	totalDist := 0.0
	for i := 0; i < bestSegmentIndex; i++ {
		start := r.Points[i]
		end := r.Points[i+1]
		segmentDist := calculateDistance(start.Latitude, start.Longitude, end.Latitude, end.Longitude)
		totalDist += segmentDist
	}
	totalDist += bestDistanceAlong

	return bestPoint, bestSegmentIndex, bestDistanceAlong, totalDist
}

// GetPointAtDistance returns a point at a specific distance along the route
func (r *Route) GetPointAtDistance(startSegmentIndex int, startDistanceAlong, targetDistance float64) RoutePoint {
	remainingDistance := targetDistance

	// Start from the given segment
	for i := startSegmentIndex; i < len(r.Points)-1; i++ {
		start := r.Points[i]
		end := r.Points[i+1]
		segmentLength := calculateDistance(start.Latitude, start.Longitude, end.Latitude, end.Longitude)

		// If starting from middle of segment, adjust segment length
		availableLength := segmentLength
		if i == startSegmentIndex {
			availableLength = segmentLength - startDistanceAlong
		}

		if remainingDistance <= availableLength {
			// Target point is on this segment
			ratio := remainingDistance / segmentLength
			if i == startSegmentIndex {
				// Adjust for starting position within segment
				ratio = (startDistanceAlong + remainingDistance) / segmentLength
			}

			// Interpolate between start and end
			lat := start.Latitude + (end.Latitude-start.Latitude)*ratio
			lng := start.Longitude + (end.Longitude-start.Longitude)*ratio

			return RoutePoint{Latitude: lat, Longitude: lng}
		}

		remainingDistance -= availableLength
	}

	// If we've reached the end, return the last point
	if len(r.Points) > 0 {
		return r.Points[len(r.Points)-1]
	}

	return RoutePoint{}
}

// closestPointOnSegment finds the closest point on a line segment to a given point
func closestPointOnSegment(lat, lng float64, start, end RoutePoint) (RoutePoint, float64) {
	// Convert to meters for easier calculation
	A := RoutePoint{Latitude: lat, Longitude: lng}
	B := start
	C := end

	// Vector from B to C
	BCx := C.Longitude - B.Longitude
	BCy := C.Latitude - B.Latitude

	// Vector from B to A
	BAx := A.Longitude - B.Longitude
	BAy := A.Latitude - B.Latitude

	// Calculate the projection
	dotProduct := BAx*BCx + BAy*BCy
	lengthSquared := BCx*BCx + BCy*BCy

	if lengthSquared == 0 {
		// B and C are the same point
		return start, 0
	}

	t := dotProduct / lengthSquared

	// Clamp t to [0, 1] to stay on the segment
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	// Calculate the closest point
	closestLat := B.Latitude + t*(C.Latitude-B.Latitude)
	closestLng := B.Longitude + t*(C.Longitude-B.Longitude)

	// Calculate distance along segment
	segmentLength := calculateDistance(B.Latitude, B.Longitude, C.Latitude, C.Longitude)
	distanceAlong := t * segmentLength

	return RoutePoint{Latitude: closestLat, Longitude: closestLng}, distanceAlong
}

// FindBestRoute determines which route a bus is most likely following based on its current position and color
func (rm *RouteMatcher) FindBestRoute(lat, lng float64, routeColor string) *Route {
	// First try to match by color
	if route, exists := rm.Routes[routeColor]; exists && routeColor != "grey" {
		return route
	}

	// If no color match or grey color, find closest route by distance
	minDistance := math.Inf(1)
	var bestRoute *Route

	for _, route := range rm.Routes {
		_, _, _, distance := route.FindClosestRoutePoint(lat, lng)
		if distance < minDistance {
			minDistance = distance
			bestRoute = route
		}
	}

	return bestRoute
}
