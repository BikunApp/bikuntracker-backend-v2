package bus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
	"github.com/FreeJ1nG/bikuntracker-backend/utils"
	"github.com/gammazero/deque"
)

const (
	DQ_SIZE = 10
)

type container struct {
	config         *utils.Config
	rmService      interfaces.RMService
	damriService   interfaces.DamriService
	busCoordinates map[string]*models.BusCoordinate
	storedBuses    map[string]*deque.Deque[*models.BusCoordinate]
}

func NewContainer(config *utils.Config, rmService interfaces.RMService, damriService interfaces.DamriService) *container {
	return &container{
		config:         config,
		rmService:      rmService,
		damriService:   damriService,
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
		if c.storedBuses[imei].Len() < DQ_SIZE {
			c.storedBuses[imei].PushBack(bus)
		} else {
			c.storedBuses[imei].PopFront()
			c.storedBuses[imei].PushBack(bus)
		}
	}
}

func (c *container) GetBusCoordinates() (res []models.BusCoordinate) {
	res = make([]models.BusCoordinate, 0)
	for _, busCoordinate := range c.busCoordinates {
		res = append(res, *busCoordinate)
	}
	return
}

func (c *container) RunChangeLaneCron() {
	for {
		time.Sleep(time.Millisecond * 5500 * DQ_SIZE)

		data := make(map[string][]*models.BusCoordinate)

		for imei, dq := range c.storedBuses {
			if dq.Len() == DQ_SIZE {
				lastFewPoints := make([]*models.BusCoordinate, 0)
				for i := 0; i < dq.Len(); i++ {
					lastFewPoints = append(lastFewPoints, dq.At(i))
				}
				data[imei] = lastFewPoints
			}
		}

		log.Println(" ::", data)

		res, err := c.rmService.DetectLane(data)
		if err != nil {
			log.Printf("rmService.DetectLane(): %s\n", err.Error())
			continue
		}

		fmt.Println(" >>", res)

		for imei, state := range res {
			fmt.Println(" >>", imei, state)
		}
	}
}

func (c *container) RunCron() {
	for {
		time.Sleep(time.Millisecond * 5500)

		busStatuses, err := c.damriService.GetAllBusStatus()
		if err != nil {
			log.Printf("damriService.GetAllBusStatus(): %s\n", err.Error())
			continue
		}

		imeiList := make([]string, 0)
		for _, busStatus := range busStatuses {
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
			for _, busStatus := range busStatuses {
				if busStatus.Imei != coordinates[imei].Imei {
					continue
				}
				coordinates[imei].Color = busStatus.Color
				coordinates[imei].Id = busStatus.BusId
			}
		}

		c.insertFetchedData(coordinates)

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
