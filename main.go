package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/FreeJ1nG/bikuntracker-backend/app/bus"
	"github.com/FreeJ1nG/bikuntracker-backend/app/damri"
	"github.com/FreeJ1nG/bikuntracker-backend/utils"
	"github.com/coder/websocket"
)

func main() {
	config, err := utils.SetupConfig()
	if err != nil {
		log.Fatalf(err.Error())
		return
	}

	damriService := damri.NewService(config)
	busContainer := bus.NewContainer(config, damriService)

	go busContainer.RunCron()

	http.HandleFunc("/v2", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)

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
			message, err := json.Marshal(coordinates)
			if err != nil {
				reason = fmt.Sprintf("unable to marshal bus coordinates: %s", err.Error())
				log.Println(reason)
				c.Close(websocket.StatusAbnormalClosure, reason)
				return
			}

			c.Write(r.Context(), websocket.MessageText, message)
			time.Sleep(time.Second * 3)
		}
	}))

	fmt.Printf("Listening on port %s ...\n", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}
