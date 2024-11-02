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
	"github.com/FreeJ1nG/bikuntracker-backend/utils"
	"github.com/gammazero/deque"
)

const (
	DQ_SIZE = 50
)

type container struct {
	config         *utils.Config
	rmService      interfaces.RMService
	damriService   interfaces.DamriService
	busService     interfaces.BusService
	busCoordinates map[string]*models.BusCoordinate
	storedBuses    map[string]*deque.Deque[*models.BusCoordinate]
}

func NewContainer(
	config *utils.Config,
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
		storedBuses:    make(map[string]*deque.Deque[*models.BusCoordinate]),
	}
}

func (c *container) insertFetchedData(buses map[string]*models.BusCoordinate) {
	for imei, bus := range buses {
		_, ok := c.storedBuses[imei]
		if !ok {
			c.storedBuses[imei] = deque.New[*models.BusCoordinate]()
		}
		if c.storedBuses[imei].Len() > 0 {
			back := c.storedBuses[imei].Back()
			if bus.Latitude == back.Latitude && bus.Longitude == back.Longitude {
				// If the data is a duplicate, we don't care about it
				continue
			}
		}
		if c.storedBuses[imei].Len() < DQ_SIZE {
			c.storedBuses[imei].PushBack(bus)
		} else {
			c.storedBuses[imei].PopFront()
			c.storedBuses[imei].PushBack(bus)
		}
	}
}

func (c *container) possiblyChangeBusLane() (err error) {
	data := make(map[string][]*models.BusCoordinate)

	for imei, dq := range c.storedBuses {
		if dq.Len() == DQ_SIZE {
			lastFewPoints := make([]*models.BusCoordinate, 0)
			for i := 0; i < dq.Len(); i++ {
				lastFewPoints = append(lastFewPoints, dq.At(i))
			}
			data[imei] = lastFewPoints
			log.Printf("Calling detect lane for bus %v", imei)
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
					_, err := c.busService.UpdateBusColorByImei(ctx, imei, cleanedColor)
					if err != nil {
						log.Printf("Unable to update bus color by imei of %s to %s", imei, state)
					}
				}
			}
		}
	}

	return
}

func (c *container) GetBusCoordinates() (res []models.BusCoordinate) {
	res = make([]models.BusCoordinate, 0)
	for _, busCoordinate := range c.busCoordinates {
		res = append(res, *busCoordinate)
	}
	return
}

func (c *container) RunCron() {
	counter := 0

	for {
		time.Sleep(time.Millisecond * 5500)

		ctx := context.Background()
		buses, err := c.busService.GetAllBuses(ctx)
		if err != nil {
			log.Printf("damriService.GetAllBusStatus(): %s\n", err.Error())
			continue
		}

		imeiList := make([]string, 0)
		for _, busStatus := range buses {
			imeiList = append(imeiList, busStatus.Imei)
		}

		coordinates, err := c.damriService.GetBusCoordinates(imeiList)
		if err != nil {
			// If this fails, we try to authenticate before redoing the request
			newToken, err := c.damriService.Authenticate()
			if err != nil {
				log.Printf("damriService.Authenticate(): %s\n", err.Error())
				continue
			}
			c.config.Token = newToken
			coordinates, err = c.damriService.GetBusCoordinates(imeiList)
			if err != nil {
				log.Printf("damriService.GetBusCoordinates(): %s\n", err.Error())
				continue
			}
		}

		for imei := range coordinates {
			for _, bus := range buses {
				if bus.Imei != coordinates[imei].Imei {
					continue
				}
				coordinates[imei].Color = bus.Color
				coordinates[imei].Id = bus.Id
			}
		}

		c.insertFetchedData(coordinates)
		counter += 1
		counter %= DQ_SIZE

		if counter == 0 {
			err := c.possiblyChangeBusLane()
			if err != nil {
				log.Printf("Unable to change bus lane: %s", err.Error())
			}
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
