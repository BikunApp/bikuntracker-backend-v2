package utils

import (
	"log"

	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
	"github.com/spf13/viper"
)

func SetupConfig() (config *models.Config, err error) {
	viper.AddConfigPath(".")
	viper.AddConfigPath("../")
	viper.AddConfigPath("/work")
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		log.Printf("Failed to read .env file: %s, trying environment variables", err.Error())
		// Continue with environment variables only
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	config.SetDBString()
	config.SetInterpolationDefaults() // Set hardcoded interpolation defaults
	return config, nil
}
