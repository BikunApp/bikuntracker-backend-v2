package bus

import (
	"fmt"
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

		c.busCoordinates = coordinates

		time.Sleep(time.Millisecond * 5500)
	}
}
