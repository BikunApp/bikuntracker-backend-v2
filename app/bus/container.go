package bus

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
	"github.com/coder/websocket"
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
	halteHistory   map[string][]string // imei -> halte name history
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
		halteHistory:   make(map[string][]string),
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

func (c *container) RunWebSocket() {
	wsUrl := c.config.WsUrl
	if wsUrl == "" {
		log.Println("WS_URL is not set in config")
		return
	}

	// Set all buses to grey at runtime start
	ctx := context.Background()
	buses, err := c.busService.GetAllBuses(ctx)
	if err == nil {
		for _, bus := range buses {
			_, _ = c.busService.UpdateBusColorByImei(ctx, bus.Imei, "grey")
		}
	}

	for {
		ctx := context.Background()
		c.connectAndConsumeWS(ctx, wsUrl)
		// If connection drops, wait a bit and retry
		time.Sleep(3 * time.Second)
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

		var wsResp struct {
			Message string                   `json:"message"`
			Data    []map[string]interface{} `json:"data"`
		}
		err = json.Unmarshal(data, &wsResp)
		if err != nil {
			log.Printf("WebSocket JSON unmarshal error: %v", err)
			continue
		}

		coordinates := make(map[string]*models.BusCoordinate)
		for _, d := range wsResp.Data {
			imei, _ := d["imei"].(string)
			lat, _ := d["latitude"].(float64)
			lng, _ := d["longitude"].(float64)
			speed, _ := d["speed"].(float64)

			// Predict route from halteHistory
			history := c.halteHistory[imei]
			routeType := detectRouteColor(history)
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

			currentHalte, dist := nearestHalte(lat, lng)
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
			} else if len(history) > 0 {
				// Show previous halte with "Depart from [Halte Name]"
				prevHalte := history[len(history)-1]
				bus.CurrentHalte = prevHalte
				bus.StatusMessage = "Depart from " + prevHalte
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

		buses, err := c.busService.GetAllBuses(ctx)
		if err == nil {
			for _, bus := range buses {
				if bc, ok := coordinates[bus.Imei]; ok {
					bc.Color = bus.Color
					bc.Id = bus.Id
				}
			}
		}

		c.insertFetchedData(coordinates)

		// Track halte visits and update halteHistory for each bus
		for imei, coord := range coordinates {
			name, dist := nearestHalte(coord.Latitude, coord.Longitude)
			if name != "" && dist < 60 { // 60 meters threshold
				history := c.halteHistory[imei]
				if len(history) == 0 || history[len(history)-1] != name {
					c.halteHistory[imei] = append(history, name)
					log.Printf("Bus %s visited halte: %s", imei, name)
					// Update current halte in DB
					ctx := context.Background()
					_, err := c.busService.UpdateCurrentHalteByImei(ctx, imei, name)
					if err != nil {
						log.Printf("Failed to update current halte for %s: %v", imei, err)
					}
				}
			}
		}

		// After updating halteHistory, auto-detect and update bus color
		for imei, history := range c.halteHistory {
			color := detectRouteColor(history)
			prevColor := ""
			if c.busCoordinates[imei] != nil {
				prevColor = c.busCoordinates[imei].Color
			}
			if color == "grey" && prevColor != "" && prevColor != "grey" {
				// If ambiguous, keep previous non-grey color
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

		err = c.possiblyChangeBusLane()
		if err != nil {
			log.Printf("Unable to change bus lane: %s", err.Error())
		}

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

		c.busCoordinates = coordinates
	}
}

func (c *container) insertFetchedData(buses map[string]*models.BusCoordinate) {
	for imei, bus := range buses {
		_, ok := c.storedBuses[imei]
		if !ok {
			c.storedBuses[imei] = &dqStore{
				dq:      deque.New[*models.BusCoordinate](),
				counter: 0,
			}
		}

		if c.storedBuses[imei].dq.Len() > 0 {
			back := c.storedBuses[imei].dq.Back()
			if bus.Latitude == back.Latitude && bus.Longitude == back.Longitude {
				// If the data is a duplicate, we don't care about it
				continue
			}
		}

		c.storedBuses[imei].counter++
		c.storedBuses[imei].counter %= DQ_SIZE
		if c.storedBuses[imei].dq.Len() < DQ_SIZE {
			c.storedBuses[imei].dq.PushBack(bus)
		} else {
			c.storedBuses[imei].dq.PopFront()
			c.storedBuses[imei].dq.PushBack(bus)
		}
	}
}

func (c *container) possiblyChangeBusLane() (err error) {
	data := make(map[string][]*models.BusCoordinate)

	for imei, dqStore := range c.storedBuses {
		if dqStore.counter == 0 && dqStore.dq.Len() == DQ_SIZE {
			// We only want to process the deque when it is full of new bus data
			// This case is when the deque is full (len == DQ SIZE) AND the circular counter is back to zero
			lastFewPoints := make([]*models.BusCoordinate, 0)
			for i := 0; i < dqStore.dq.Len(); i++ {
				lastFewPoints = append(lastFewPoints, dqStore.dq.At(i))
			}
			data[imei] = lastFewPoints
			res, err := c.rmService.DetectLane(imei, lastFewPoints)
			if err != nil {
				log.Printf("Unable to detect lane for bus %v", err.Error())
				continue
			}
			for imei, state := range res {
				if state != "unknown" {
					ctx := context.Background()
					var cleanedColor string
					if state == "blue" {
						cleanedColor = "biru"
					} else if state == "red" {
						cleanedColor = "merah"
					} else {
						log.Printf("Bus color is not blue or red, something is probably wrong")
						continue
					}
					log.Printf("Detected lane for bus %v: %v", imei, cleanedColor)
					_, err := c.busService.UpdateBusColorByImei(ctx, imei, cleanedColor)
					if err != nil {
						log.Printf("Unable to update bus color by imei of %s to %s", imei, cleanedColor)
					}
				}
			}
		}
	}

	return
}

// Halte metadata and route definitions
type Halte struct {
	Name     string
	Lat, Lng float64
}

var halteList = []Halte{
	{"Asrama UI", -6.348351370044594, 106.82976588606834},
	{"Menwa", -6.353471269466313, 106.83177955448627},
	{"Stasiun UI", -6.361052900888018, 106.83170076459645},
	{"Fakultas Psikologi", -6.36255935735158, 106.83111906051636},
	{"FISIP", -6.361574, 106.830172},
	{"Fakultas Ilmu Pengetahuan Budaya", -6.361254501381427, 106.82978868484497},
	{"Fakultas Ekonomi dan Bisnis", -6.35946048561971, 106.82582974433899},
	{"Fakultas Teknik", -6.361043911445512, 106.82325214147568},
	{"Vokasi", -6.366036735678631, 106.8216535449028},
	{"SOR", -6.366915739619239, 106.82448193430899},
	{"FMIPA", -6.369828304090281, 106.8257811293006},
	{"Fakultas Ilmu Keperawatan", -6.371008186217929, 106.8268945813179},
	{"Fakultas Kesehatan Masyarakat", -6.371677262480034, 106.8293622136116},
	{"RIK", -6.36987795182555, 106.8310546875},
	{"Balairung", -6.368212251024606, 106.83178257197142},
	{"MUI/Perpus UI", -6.3655942342627565, 106.83204710483551},
	{"Fakultas Hukum", -6.364901492199248, 106.83221206068993},
}

var blueNormal = []string{"Asrama UI", "Menwa", "Stasiun UI", "Fakultas Psikologi", "FISIP", "Fakultas Ilmu Pengetahuan Budaya", "Fakultas Ekonomi dan Bisnis", "Fakultas Teknik", "Vokasi", "SOR", "FMIPA", "Fakultas Ilmu Keperawatan", "Fakultas Kesehatan Masyarakat", "RIK", "Balairung", "MUI/Perpus UI", "Fakultas Hukum", "Stasiun UI", "Menwa"}
var blueMorning = []string{"Asrama UI", "Menwa", "Stasiun UI", "Fakultas Psikologi", "FISIP", "Fakultas Ilmu Pengetahuan Budaya", "Fakultas Ekonomi dan Bisnis", "Fakultas Teknik", "Vokasi", "SOR", "FMIPA", "Fakultas Ilmu Keperawatan", "Fakultas Kesehatan Masyarakat", "RIK", "Balairung", "MUI/Perpus UI", "Fakultas Hukum"}
var redNormal = []string{"Asrama UI", "Menwa", "Stasiun UI", "Fakultas Hukum", "Balairung", "RIK", "Fakultas Kesehatan Masyarakat", "Fakultas Ilmu Keperawatan", "FMIPA", "SOR", "Vokasi", "Fakultas Teknik", "Fakultas Ekonomi dan Bisnis", "Fakultas Ilmu Pengetahuan Budaya", "FISIP", "Fakultas Psikologi", "Stasiun UI", "Menwa"}
var redMorning = []string{"Asrama UI", "Menwa", "Stasiun UI", "Fakultas Hukum", "Balairung", "RIK", "Fakultas Kesehatan Masyarakat", "Fakultas Ilmu Keperawatan", "FMIPA", "SOR", "Vokasi"}

// nearestHalte returns the name and distance (in meters) of the closest halte to the given latitude and longitude.
func nearestHalte(lat, lng float64) (string, float64) {
	const earthRadius = 6371000 // meters
	minDist := 1e9
	closest := ""
	for _, halte := range halteList {
		dLat := (halte.Lat - lat) * (3.141592653589793 / 180)
		dLng := (halte.Lng - lng) * (3.141592653589793 / 180)
		alat := lat * (3.141592653589793 / 180)
		blat := halte.Lat * (3.141592653589793 / 180)
		a := (dLat/2)*(dLat/2) + (dLng/2)*(dLng/2)*cos(alat)*cos(blat)
		c := 2 * atan2Sqrt(a, 1-a)
		dist := earthRadius * c
		if dist < minDist {
			minDist = dist
			closest = halte.Name
		}
	}
	return closest, minDist
}

// Helper math functions for nearestHalte
func cos(x float64) float64 {
	return float64(math.Cos(float64(x)))
}
func atan2Sqrt(a, b float64) float64 {
	return math.Atan2(math.Sqrt(a), math.Sqrt(b))
}

// detectRouteColor tries to determine the route color and type based on halte visit history.
func detectRouteColor(history []string) string {
	if len(history) < 4 {
		return "grey" // grey if not enough halte visited
	}
	seq := history[len(history)-4:]
	// Helper to check if a sequence exists in a route
	matches := func(route []string) bool {
		for i := 0; i <= len(route)-4; i++ {
			if route[i] == seq[0] && route[i+1] == seq[1] && route[i+2] == seq[2] {
				return true
			}
		}
		return false
	}
	switch {
	case matches(blueNormal):
		return "blue"
	case matches(blueMorning):
		return "express-blue"
	case matches(redNormal):
		return "red"
	case matches(redMorning):
		return "express-red"
	default:
		return "grey"
	}
}
