package utils

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	DamriApi           string `mapstructure:"DAMRI_API"`
	DamriLoginUsername string `mapstructure:"DAMRI_LOGIN_USERNAME"`
	DamriLoginPassword string `mapstructure:"DAMRI_LOGIN_PASSWORD"`

	BikunAdminApi    string `mapstructure:"BIKUN_ADMIN_API"`
	BikunAdminApiKey string `mapstructure:"BIKUN_ADMIN_API_KEY"`

	Port         string `mapstructure:"PORT"`
	PrintCsvLogs bool   `mapstructure:"PRINT_CSV_LOGS"`

	Token string
}

func SetupConfig() (config *Config, err error) {
	viper.AddConfigPath("../")
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		log.Printf("Failed to read .env file: %s", err.Error())
		return
	}

	err = viper.Unmarshal(&config)
	return
}
