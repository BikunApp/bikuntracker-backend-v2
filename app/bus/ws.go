package bus

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
	"github.com/coder/websocket"
)

func (c *container) RunWebSocket() {
	wsUrl := c.config.WsUrl
	if wsUrl == "" {
		log.Println("WS_URL is not set in config")
		return
	}

	log.Printf("Starting lap tracking system - WebSocket URL: %s", wsUrl)

	ctx := context.Background()
	buses, err := c.busService.GetAllBuses(ctx)
	if err == nil {
		log.Printf("Found %d buses in database", len(buses))
		for _, bus := range buses {
			_, _ = c.busService.UpdateBusColorByImei(ctx, bus.Imei, "grey")

			// Initialize active lap status
			activeLap, _ := c.busService.GetActiveLap(ctx, bus.Imei)
			c.activeLaps[bus.Imei] = activeLap != nil
			if activeLap != nil {
				log.Printf("Found active lap for bus %s: lap %d", bus.Imei, activeLap.LapNumber)
			}
		}
	} else {
		log.Printf("Failed to get buses: %v", err)
	}

	log.Printf("Lap detection rules - Start: Asrama UI → Menwa, End: → Parking OR → Asrama UI")

	for {
		ctx := context.Background()
		c.connectAndConsumeWS(ctx, wsUrl)
		time.Sleep(1 * time.Second)
	}
}

func (c *container) connectAndConsumeWS(ctx context.Context, wsUrl string) {
	log.Printf("Connecting to WebSocket: %s", wsUrl)
	conn, _, err := websocket.Dial(ctx, wsUrl, nil)
	if err != nil {
		log.Printf("WebSocket dial error: %v", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			return
		}
		coordinates := c.parseWSData(data)

		c.updateBusColors(coordinates)
		c.insertFetchedData(coordinates)
		c.updateHalteVisits(ctx, coordinates)
		err = c.possiblyChangeBusLane()
		if err != nil {
			log.Printf("Unable to change bus lane: %s", err.Error())
		}
		c.logCsvIfNeeded(coordinates)
		c.busCoordinates = coordinates
	}
}

func (c *container) parseWSData(data []byte) map[string]*models.BusCoordinate {
	var wsResp struct {
		Message string                   `json:"message"`
		Data    []map[string]interface{} `json:"data"`
	}
	err := json.Unmarshal(data, &wsResp)
	if err != nil {
		log.Printf("WebSocket JSON unmarshal error: %v", err)
		return nil
	}
	coordinates := make(map[string]*models.BusCoordinate)
	for _, d := range wsResp.Data {
		imei, _ := d["imei"].(string)
		lat, _ := d["latitude"].(float64)
		lng, _ := d["longitude"].(float64)
		speed, _ := d["speed"].(float64)
		currentHalte, dist := nearestHalte(lat, lng)
		previousHalte := c.previousHalte[imei]
		routeType := detectRouteColorFromPair(previousHalte, currentHalte)
		var route []string
		switch routeType {
		case "blue":
			route = blueNormal
		case "express-blue":
			route = blueMorning
		case "red":
			route = redNormal
		case "express-red":
			route = redMorning
		default:
			route = nil
		}
		bus := &models.BusCoordinate{
			Imei:         imei,
			Latitude:     lat,
			Longitude:    lng,
			Speed:        int(speed),
			GpsTime:      time.Now(),
			CurrentHalte: "",
			NextHalte:    "",
		}
		if currentHalte != "" && dist < 60 {
			bus.CurrentHalte = currentHalte
			bus.StatusMessage = "Arriving at " + currentHalte
		} else if previousHalte != "" {
			bus.CurrentHalte = previousHalte
			bus.StatusMessage = "Depart from " + previousHalte
		}
		if route != nil && bus.CurrentHalte != "" {
			for i, h := range route {
				if h == bus.CurrentHalte {
					if i+1 < len(route) {
						bus.NextHalte = route[i+1]
					} else if len(route) > 0 {
						bus.NextHalte = route[0]
					}
					break
				}
			}
		}
		coordinates[imei] = bus
	}
	buses, err := c.busService.GetAllBuses(context.Background())
	if err == nil {
		for _, bus := range buses {
			if bc, ok := coordinates[bus.Imei]; ok {
				bc.Color = bus.Color
				bc.Id = bus.Id
			}
		}
	}
	return coordinates
}

func (c *container) updateBusColors(coordinates map[string]*models.BusCoordinate) {
	for imei, coord := range coordinates {
		name, dist := nearestHalte(coord.Latitude, coord.Longitude)
		if name != "" && dist < 60 {
			previousHalte := c.previousHalte[imei]
			color := detectRouteColorFromPair(previousHalte, name)
			prevColor := ""
			if c.busCoordinates[imei] != nil {
				prevColor = c.busCoordinates[imei].Color
			}
			if color == "grey" && prevColor != "" && prevColor != "grey" {
				continue
			}
			if c.busCoordinates[imei] != nil && c.busCoordinates[imei].Color != color {
				c.busCoordinates[imei].Color = color
				ctx := context.Background()
				_, err := c.busService.UpdateBusColorByImei(ctx, imei, color)
				if err != nil {
					log.Printf("Failed to update bus color for %s: %v", imei, err)
				} else {
					log.Printf("Auto-detected and updated bus %s color to %s", imei, color)
				}
			}
		}
	}
}

func (c *container) updateHalteVisits(ctx context.Context, coordinates map[string]*models.BusCoordinate) {
	for imei, coord := range coordinates {
		name, dist := nearestHalte(coord.Latitude, coord.Longitude)
		if name != "" && dist < 60 {
			currentPrevious := c.previousHalte[imei]
			if currentPrevious != name {
				log.Printf("Bus %s halte switch: %s → %s (%.1fm)", imei, currentPrevious, name, dist)

				// Track halte visit for active lap (before checking lap start/end conditions)
				if c.activeLaps[imei] {
					err := c.busService.AddHalteVisitToActiveLap(ctx, imei, name)
					if err != nil {
						log.Printf("Failed to add halte visit to active lap for bus %s: %v", imei, err)
					} else {
						log.Printf("Added halte visit '%s' to active lap for bus %s", name, imei)
					}
				}

				// Check for lap start: transition from "Asrama UI" to "Menwa"
				if currentPrevious == "Asrama UI" && name == "Menwa" {
					log.Printf("Lap start condition met - Bus %s: Asrama UI → Menwa", imei)

					routeColor := coord.Color
					if routeColor == "" {
						routeColor = "grey"
					}

					// End existing lap if one is active before starting new one
					if c.activeLaps[imei] {
						log.Printf("Ending previous lap for bus %s to start new one", imei)
						_, err := c.busService.EndLap(ctx, imei)
						if err != nil {
							log.Printf("Failed to end previous lap for bus %s: %v", imei, err)
						}
					}

					lapHistory, err := c.busService.StartLap(ctx, imei, routeColor)
					if err != nil {
						log.Printf("Failed to start lap for bus %s: %v", imei, err)
					} else {
						c.activeLaps[imei] = true
						log.Printf("Started lap %d for bus %s (color: %s)", lapHistory.LapNumber, imei, routeColor)
						c.pushLapEvent(ctx, imei, "lap_start", lapHistory)
					}
				}

				// Check for lap end: reaching "Parking" or returning to "Asrama UI" (if coming from elsewhere)
				if c.activeLaps[imei] && (name == "Parking" || (name == "Asrama UI" && currentPrevious == "Menwa")) {
					log.Printf("Lap end condition met - Bus %s reached %s from %s", imei, name, currentPrevious)

					lapHistory, err := c.busService.EndLap(ctx, imei)
					if err != nil {
						log.Printf("Failed to end lap for bus %s: %v", imei, err)
					} else if lapHistory != nil {
						c.activeLaps[imei] = false
						log.Printf("Ended lap %d for bus %s", lapHistory.LapNumber, imei)
						c.pushLapEvent(ctx, imei, "lap_end", lapHistory)
					}
				}

				// Now update the previous halte AFTER checking lap conditions
				c.previousHalte[imei] = name

				_, err := c.busService.UpdateCurrentHalteByImei(ctx, imei, name)
				if err != nil {
					log.Printf("Failed to update current halte for %s: %v", imei, err)
				}
			}
		}
	}
}

func (c *container) logCsvIfNeeded(coordinates map[string]*models.BusCoordinate) {
	if c.config.PrintCsvLogs {
		body, err := json.Marshal(map[string]interface{}{
			"coordinates": coordinates,
		})
		if err != nil {
			log.Printf("unable to upload logs: %s", err.Error())
		} else {
			resp, err := http.Post("http://localhost:4040", "application/json", bytes.NewBuffer(body))
			if err != nil || resp.StatusCode < 200 && resp.StatusCode >= 300 {
				log.Printf("something went wrong when trying to POST logs: %s", err.Error())
			}
		}
	}
}

func (c *container) pushLapEvent(ctx context.Context, imei string, eventType string, lapHistory *models.BusLapHistory) {
	// Create lap event data using DTO structure
	eventData := dto.LapEventData{
		EventType:         eventType,
		IMEI:              imei,
		LapID:             lapHistory.ID,
		LapNumber:         lapHistory.LapNumber,
		RouteColor:        lapHistory.RouteColor,
		HalteVisitHistory: lapHistory.HalteVisitHistory,
		StartTime:         lapHistory.StartTime,
		Timestamp:         time.Now(),
	}

	if lapHistory.EndTime != nil {
		eventData.EndTime = lapHistory.EndTime
		duration := lapHistory.EndTime.Sub(lapHistory.StartTime).Seconds()
		eventData.Duration = &duration
	}

	// Convert to JSON for logging/pushing
	eventJSON, err := json.Marshal(eventData)
	if err != nil {
		log.Printf("Failed to marshal lap event data: %v", err)
		return
	}

	log.Printf("Lap event: %s", string(eventJSON))

	// Here you can add additional logic to push data to external systems
	// For example, HTTP POST to an external API, WebSocket broadcast, etc.
	if c.config.PrintCsvLogs {
		resp, err := http.Post("http://localhost:4040/lap-events", "application/json", bytes.NewBuffer(eventJSON))
		if err != nil {
			log.Printf("Failed to push lap event: %v", err)
		} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Printf("Successfully pushed %s event for bus %s", eventType, imei)
		}
	}
}
