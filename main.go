package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

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

	damriService := damri.NewService(&config)

	http.HandleFunc("/v2", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reason string
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			reason = fmt.Sprintf("unable to upgrade websocket connection: %s", err.Error())
			fmt.Println(reason)
			c.Close(websocket.StatusNormalClosure, reason)
			return
		}
		defer c.CloseNow()

		for {
			busStatuses, err := damriService.GetAllBusStatus()
			if err != nil {
				reason = fmt.Sprintf("damriService.GetAllBusStatus(): %s", err.Error())
				fmt.Println(reason)
				c.Close(websocket.StatusNormalClosure, reason)
				return
			}

			imeiList := make([]string, 0)
			for _, busStatus := range busStatuses {
				imeiList = append(imeiList, busStatus.Imei)
			}

			coordinates, err := damriService.GetBusCoordinates(imeiList)
			if err != nil {
				// If this fails, we try to authenticate before redoing the request
				newToken, err := damriService.Authenticate()
				if err != nil {
					reason = fmt.Sprintf("damriService.Authenticate(): %s", err.Error())
					fmt.Println(reason)
					c.Close(websocket.StatusAbnormalClosure, reason)
					return
				}
				config.Token = newToken
				coordinates, err = damriService.GetBusCoordinates(imeiList)
				if err != nil {
					reason = fmt.Sprintf("damriService.GetBusCoordinates(): %s", err.Error())
					fmt.Println(reason)
					c.Close(websocket.StatusAbnormalClosure, reason)
					return
				}
			}

			message, err := json.Marshal(coordinates)
			if err != nil {
				reason = fmt.Sprintf("unable to marshal bus coordinates: %s", err.Error())
				fmt.Println(reason)
				c.Close(websocket.StatusAbnormalClosure, reason)
				return
			}

			c.Write(r.Context(), websocket.MessageText, message)
			time.Sleep(time.Second * 4)
		}
	}))

	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}
