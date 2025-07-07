package bus

import (
	"context"
	"log"

	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
)

func updateHalteHistoryAndDB(ctx context.Context, coordinates map[string]*models.BusCoordinate, halteHistory map[string][]string, busService interfaces.BusService) {
	for imei, coord := range coordinates {
		name, dist := nearestHalte(coord.Latitude, coord.Longitude)
		if name != "" && dist < 60 {
			history := halteHistory[imei]
			if len(history) == 0 || history[len(history)-1] != name {
				halteHistory[imei] = append(history, name)
				if len(halteHistory[imei]) > 2 {
					halteHistory[imei] = halteHistory[imei][len(halteHistory[imei])-2:]
				}
				log.Printf("Bus %s visited halte: %s", imei, name)
				_, err := busService.UpdateCurrentHalteByImei(ctx, imei, name)
				if err != nil {
					log.Printf("Failed to update current halte for %s: %v", imei, err)
				}
			}
		}
	}
}
