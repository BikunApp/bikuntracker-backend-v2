package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/FreeJ1nG/bikuntracker-backend/app/bus"
	"github.com/FreeJ1nG/bikuntracker-backend/app/damri"
	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/utils"
	"github.com/coder/websocket"
)

func main() {
	config, err := utils.SetupConfig()
	if err != nil {
		log.Fatalf(err.Error())
		return
	}

	damriUtil := damri.NewUtil()
	damriService := damri.NewService(config, damriUtil)
	busContainer := bus.NewContainer(config, damriService)

	go busContainer.RunCron()

	http.HandleFunc("/v2", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{"localhost:5173"},
		})

		var reason string
		if err != nil {
			reason = fmt.Sprintf("unable to upgrade websocket connection: %s", err.Error())
			log.Println(reason)
			c.Close(websocket.StatusAbnormalClosure, reason)
			return
		}
		defer c.CloseNow()

		for {
			coordinates := busContainer.GetBusCoordinates()

			operationalStatus, err := damriService.GetOperationalStatus()
			if err != nil {
				reason = fmt.Sprintf("damriService.GetOperationalStatus(): %s", err.Error())
				log.Println(reason)
				c.Close(websocket.StatusAbnormalClosure, reason)
				continue
			}

			message, err := json.Marshal(dto.CoordinateBroadcastMessage{Coordinates: coordinates, OperationalStatus: operationalStatus})
			if err != nil {
				reason = fmt.Sprintf("unable to marshal bus coordinates: %s", err.Error())
				log.Println(reason)
				c.Close(websocket.StatusAbnormalClosure, reason)
				continue
			}

			c.Write(r.Context(), websocket.MessageText, message)
			time.Sleep(time.Second * 3)
		}
	}))

	fmt.Printf("Listening on port %s ...\n", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}
