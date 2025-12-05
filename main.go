package main

import (
	"context"
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

	rmService := rm.NewService(config)

	damriUtil := damri.NewUtil()
	damriService := damri.NewService(config, damriUtil)
	busContainer := bus.NewContainer(config, rmService, damriService, busService)

	busHandler := bus.NewHandler(busRepo, busService, busContainer)

	// Initialize runtime caches; location updates now come via webhook instead of WS
	busContainer.InitRuntimeState()

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

	// Lap history routes
	utils.HandleRoute("/bus/lap-history", utils.MethodHandler{http.MethodGet: busHandler.GetFilteredLapHistory}, &utils.Options{
		Middlewares: []middleware.Middleware{
			adminApiKeyProtectorMiddleware,
		},
	})
	utils.HandleRoute("/bus/:imei/lap-history", utils.MethodHandler{http.MethodGet: busHandler.GetLapHistory}, nil)
	utils.HandleRoute("/bus/:imei/active-lap", utils.MethodHandler{http.MethodGet: busHandler.GetActiveLap}, nil)
	// Debug route - remove in production
	utils.HandleRoute("/bus/test-lap-data", utils.MethodHandler{http.MethodPost: busHandler.CreateTestLapData}, nil)
	utils.HandleRoute("/bus/check-table", utils.MethodHandler{http.MethodGet: busHandler.CheckLapHistoryTable}, nil)

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

			if err != nil {
				log.Printf("WebSocket upgrade failed: %v", err)
				return
			}
			defer c.CloseNow()

			// Create context with timeout for this connection
			ctx, cancel := context.WithCancel(r.Context())
			defer cancel()

			// Cache for message to avoid repeated marshaling
			var lastMessage []byte
			var lastUpdate time.Time
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			// Ping ticker to detect disconnected clients
			pingTicker := time.NewTicker(30 * time.Second)
			defer pingTicker.Stop()

			// Set read timeout to detect dead connections
			c.SetReadLimit(1024)

			// Start goroutine to handle pings and detect disconnections
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case <-pingTicker.C:
						// Send ping to detect if client is still connected
						if err := c.Ping(ctx); err != nil {
							cancel() // This will close the main loop
							return
						}
					}
				}
			}()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					// Only fetch data if we haven't updated recently or if data might be stale
					now := time.Now()
					if now.Sub(lastUpdate) < 500*time.Millisecond {
						// Send cached message if recent
						if lastMessage != nil {
							writeCtx, writeCancel := context.WithTimeout(ctx, 5*time.Second)
							if err := c.Write(writeCtx, websocket.MessageText, lastMessage); err != nil {
								writeCancel()
								return // Client disconnected or write timeout
							}
							writeCancel()
						}
						continue
					}

					// Fetch fresh data
					coordinates := busContainer.GetBusCoordinates()
					coordinatesMap := busContainer.GetBusCoordinatesMap()

					// Get operational status with timeout
					operationalStatus, err := damriService.GetOperationalStatus(coordinatesMap)
					if err != nil {
						// Log error but don't break connection - send data without operational status
						log.Printf("Warning: Failed to get operational status: %v", err)
						operationalStatus = make(map[string]interface{}) // Empty status
					}

					// Marshal message
					message, err := json.Marshal(dto.CoordinateBroadcastMessage{
						Coordinates:       coordinates,
						OperationalStatus: operationalStatus,
					})
					if err != nil {
						log.Printf("JSON marshal error: %v", err)
						continue
					}

					// Cache the message and timestamp
					lastMessage = message
					lastUpdate = now

					// Send message with write timeout
					writeCtx, writeCancel := context.WithTimeout(ctx, 5*time.Second)
					if err := c.Write(writeCtx, websocket.MessageText, message); err != nil {
						writeCancel()
						return // Client disconnected or write timeout
					}
					writeCancel()
				}
			}
		}),
		nil,
	)

	// Webhook to receive location updates
	utils.HandleRoute("/wh", utils.MethodHandler{http.MethodPost: busHandler.WebhookUpdate}, nil)

	fmt.Printf("Listening on port %s ...\n", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}
