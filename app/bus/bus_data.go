package bus

import (
	"context"
	"log"

	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
	"github.com/gammazero/deque"
)

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
