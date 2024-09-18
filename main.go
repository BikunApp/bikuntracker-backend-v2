package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
	authHandler := auth.NewHandler(authService, authRepo)

	utils.HandleRoute("/auth/sso/login", authHandler.SsoLogin, &utils.Options{AllowedMethods: []string{http.MethodPost}})
	utils.HandleRoute("/auth/refresh", authHandler.RefreshJwt, &utils.Options{AllowedMethods: []string{http.MethodPost}})
	utils.HandleRoute("/auth/me", authHandler.GetCurrentUser, &utils.Options{
		Middlewares:    []utils.Middleware{utils.JwtMiddlewareFactory(authUtil)},
		AllowedMethods: []string{http.MethodGet},
	})

	utils.HandleRoute("/",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
				OriginPatterns: strings.Split(config.WsUpgradeWhitelist, ","),
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
