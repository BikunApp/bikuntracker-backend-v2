package bus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
	"github.com/FreeJ1nG/bikuntracker-backend/utils"
)

type container struct {
	config         *utils.Config
	damriService   interfaces.DamriService
	busCoordinates []dto.BusCoordinate
}

func NewContainer(config *utils.Config, damriService interfaces.DamriService) *container {
	return &container{
		config:         config,
		damriService:   damriService,
		busCoordinates: []dto.BusCoordinate{},
	}
}

func (c *container) GetBusCoordinates() []dto.BusCoordinate {
	return c.busCoordinates
}

func (c *container) RunCron() {
	for {
		time.Sleep(time.Millisecond * 5500)

		busStatuses, err := c.damriService.GetAllBusStatus()
		if err != nil {
			fmt.Printf("damriService.GetAllBusStatus(): %s\n", err.Error())
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
				fmt.Printf("damriService.Authenticate(): %s\n", err.Error())
				continue
			}
			c.config.Token = newToken
			coordinates, err = c.damriService.GetBusCoordinates(imeiList)
			if err != nil {
				fmt.Printf("damriService.GetBusCoordinates(): %s\n", err.Error())
				continue
			}
		}

		for i := range coordinates {
			for _, busStatus := range busStatuses {
				if busStatus.Imei != coordinates[i].Imei {
					continue
				}
				coordinates[i].Color = busStatus.Color
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
