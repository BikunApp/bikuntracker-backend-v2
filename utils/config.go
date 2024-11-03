package utils

import (
	"log"

	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
	"github.com/spf13/viper"
)

func SetupConfig() (config *models.Config, err error) {
	viper.AddConfigPath("../")
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		log.Printf("Failed to read .env file: %s", err.Error())
		return
	}

	err = viper.Unmarshal(&config)
	config.SetDBString()
	return
}
