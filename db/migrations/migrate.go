package main

import (
	"flag"
	"log"

	"github.com/FreeJ1nG/bikuntracker-backend/utils"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var action string
var steps uint

func init() {
	flag.StringVar(&action, "action", "up", "run db migrations [up | down]")
	flag.UintVar(&steps, "steps", 0, "amount of migrations run. If not specified, run all")
	flag.Parse()
}

func main() {
	config, err := utils.SetupConfig()
	if err != nil {
		log.Fatal("Failed to load config: ", err.Error())
	}

	migrate, err := migrate.New("file://db/migrations", config.DBUrl)
	if err != nil {
		log.Fatal("Failed to read migration files: ", err.Error())
	}

	if action == "up" {
		if steps != 0 {
			if err = migrate.Steps(int(steps)); err != nil {
				if err.Error() == "no change" {
					log.Println("Nothing to run")
				} else {
					log.Fatal("Failed to run migration: ", err.Error())
				}
			}
		} else {
			if err = migrate.Up(); err != nil {
				if err.Error() == "no change" {
					log.Println("Nothing to run")
				} else {
					log.Fatal("Failed to run migration: ", err.Error())
				}
			}
		}
	} else if action == "down" {
		if steps != 0 {
			if err = migrate.Steps(-1 * int(steps)); err != nil {
				if err.Error() == "no change" {
					log.Println("Nothing to run")
				} else {
					log.Fatal("Failed to run migration: ", err.Error())
				}
			}
		} else {
			if err = migrate.Down(); err != nil {
				if err.Error() == "no change" {
					log.Println("Nothing to run")
				} else {
					log.Fatal("Failed to run migration: ", err.Error())
				}
			}
		}
	} else {
		log.Fatal("Invalid migration action")
	}
}
