package main

import (
	"context"
	"log"
	"os"

	"github.com/FreeJ1nG/bikuntracker-backend/app/bus"
	"github.com/FreeJ1nG/bikuntracker-backend/db"
	"github.com/FreeJ1nG/bikuntracker-backend/utils"
)

const (
	USAGE_STRING = "Usage: go run scripts/seeder/main.go [bus]."
)

func main() {
	if len(os.Args) < 2 {
		log.Println(USAGE_STRING)
		return
	}

	config, err := utils.SetupConfig()
	if err != nil {
		panic(err)
	}

	pool := db.CreatePool(config.DBDsn)
	db.TestConnection(pool)

	busRepo := bus.NewRepository(pool)
	busSeeder := bus.NewSeeder(busRepo)

	ctx := context.Background()

	switch os.Args[1] {
	case "bus":
		busSeeder.SeedBuses(ctx)
	}
}
