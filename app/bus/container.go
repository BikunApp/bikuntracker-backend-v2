package bus

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
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

// func (c *container) RunCron() {
// 	for {
// 		time.Sleep(time.Millisecond * 5500)

// 		ctx := context.Background()
// 		buses, err := c.busService.GetAllBuses(ctx)
// 		if err != nil {
// 			log.Printf("damriService.GetAllBusStatus(): %s\n", err.Error())
// 			continue
// 		}

// 		imeiList := make([]string, 0)
// 		for _, busStatus := range buses {
// 			imeiList = append(imeiList, busStatus.Imei)
// 		}

// 		coordinates, err := c.damriService.GetBusCoordinates(imeiList)
// 		if err != nil {
// 			// If this fails, we try to authenticate before redoing the request
// 			newToken, err := c.damriService.Authenticate()
// 			if err != nil {
// 				log.Printf("damriService.Authenticate(): %s\n", err.Error())
// 				continue
// 			}
// 			c.config.Token = newToken
// 			coordinates, err = c.damriService.GetBusCoordinates(imeiList)
// 			if err != nil {
// 				log.Printf("damriService.GetBusCoordinates(): %s\n", err.Error())
// 				continue
// 			}
// 		}

// 		for imei := range coordinates {
// 			for _, bus := range buses {
// 				if bus.Imei != coordinates[imei].Imei {
// 					continue
// 				}
// 				coordinates[imei].Color = bus.Color
// 				coordinates[imei].Id = bus.Id
// 			}
// 		}

// 		c.insertFetchedData(coordinates)

// 		err = c.possiblyChangeBusLane()
// 		if err != nil {
// 			log.Printf("Unable to change bus lane: %s", err.Error())
// 		}

// 		if c.config.PrintCsvLogs {
// 			body, err := json.Marshal(map[string]interface{}{
// 				"coordinates": coordinates,
// 			})
// 			if err != nil {
// 				log.Printf("unable to upload logs: %s", err.Error())
// 			} else {
// 				resp, err := http.Post("http://localhost:4040", "application/json", bytes.NewBuffer(body))
// 				if err != nil || resp.StatusCode < 200 && resp.StatusCode >= 300 {
// 					log.Printf("something went wrong when trying to POST logs: %s", err.Error())
// 				}
// 			}
// 		}

// 		c.busCoordinates = coordinates
// 	}
// }

func (c *container) RunWebSocket() {
	wsUrl := c.config.WsUrl
	if wsUrl == "" {
		log.Println("WS_URL is not set in config")
		return
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
			bus := &models.BusCoordinate{
				Imei:      imei,
				Latitude:  lat,
				Longitude: lng,
				Speed:     int(speed),
			}
			coordinates[imei] = bus
		}

		// Optionally, enrich with bus info (color, id, etc)
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
		log.Printf("Inserted %d bus coordinates", len(coordinates))

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
