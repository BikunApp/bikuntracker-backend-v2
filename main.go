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
	"github.com/FreeJ1nG/bikuntracker-backend/app/rm"
	"github.com/FreeJ1nG/bikuntracker-backend/db"
	"github.com/FreeJ1nG/bikuntracker-backend/utils"
	"github.com/FreeJ1nG/bikuntracker-backend/utils/middleware"
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

	busRepo := bus.NewRepository(pool)
	busService := bus.NewService(busRepo)
	busHandler := bus.NewHandler(busRepo)

	rmService := rm.NewService(config)

	damriUtil := damri.NewUtil()
	damriService := damri.NewService(config, damriUtil)
	busContainer := bus.NewContainer(config, rmService, damriService, busService)

	go busContainer.RunWebSocket()

	authUtil := auth.NewUtil(config)
	authRepo := auth.NewRepository(pool)
	authService := auth.NewService(authUtil, authRepo)
	authHandler := auth.NewHandler(authService, authRepo)

	roleProtectMiddlewareFactory := middleware.NewRoleProtectMiddlewareFactory(config, authRepo)
	adminApiKeyProtectorMiddleware := roleProtectMiddlewareFactory.MakeAdminApiKeyProtector()
	jwtMiddleware := middleware.NewJwtMiddlewareFactory(authUtil).Make()

	utils.HandleRoute("/bus", utils.MethodHandler{http.MethodGet: busHandler.GetBuses, http.MethodPost: busHandler.CreateBus}, &utils.Options{
		MethodSpecificMiddlewares: utils.MethodSpecificMiddlewares{
			http.MethodPost: []middleware.Middleware{
				adminApiKeyProtectorMiddleware,
			},
		},
	})
	utils.HandleRoute("/bus/:id", utils.MethodHandler{http.MethodPut: busHandler.UpdateBus, http.MethodDelete: busHandler.DeleteBus}, &utils.Options{
		Middlewares: []middleware.Middleware{
			adminApiKeyProtectorMiddleware,
		},
	})

	utils.HandleRoute("/auth/sso/login", utils.MethodHandler{http.MethodPost: authHandler.SsoLogin}, nil)
	utils.HandleRoute("/auth/refresh", utils.MethodHandler{http.MethodPost: authHandler.RefreshJwt}, nil)
	utils.HandleRoute("/auth/me",
		utils.MethodHandler{http.MethodGet: authHandler.GetCurrentUser},
		&utils.Options{Middlewares: []middleware.Middleware{jwtMiddleware}},
	)

	utils.HandleRoute("/ws",
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

				operationalStatus, err := damriService.GetOperationalStatus(busContainer.GetBusCoordinatesMap())
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
				time.Sleep(time.Second * 1)
			}
		}),
		nil,
	)

	fmt.Printf("Listening on port %s ...\n", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}
