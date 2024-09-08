package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/FreeJ1nG/bikuntracker-backend/app/auth"
	"github.com/FreeJ1nG/bikuntracker-backend/app/bus"
	"github.com/FreeJ1nG/bikuntracker-backend/app/damri"
	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/db"
	"github.com/FreeJ1nG/bikuntracker-backend/utils"
	"github.com/coder/websocket"
)

func main() {
	config, err := utils.SetupConfig()
	if err != nil {
		log.Fatalf(err.Error())
		return
	}

	pool := db.CreatePool(config.DBDsn)
	db.TestConnection(pool)

	damriUtil := damri.NewUtil()
	damriService := damri.NewService(config, damriUtil)
	busContainer := bus.NewContainer(config, damriService)

	go busContainer.RunCron()

	authUtil := auth.NewUtil(config)
	authRepo := auth.NewRepository(pool)
	authService := auth.NewService(authUtil, authRepo)

	utils.HandleRoute(
		"/sso/login",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := utils.ParseRequestBody[dto.SSOLoginRequestBody](r.Body)

			accessToken, refreshToken, err := authService.SSOLogin(body.Ticket, body.Service)
			if err != nil {
				log.Printf(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			response, err := json.Marshal(dto.TokenResponse{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			})
			if err != nil {
				log.Printf(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fmt.Println(" >>", string(response))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(response)

			if err != nil {
				log.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}),
		nil,
	)

	utils.HandleRoute("/",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		}),
		nil,
	)

	fmt.Printf("Listening on port %s ...\n", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}
