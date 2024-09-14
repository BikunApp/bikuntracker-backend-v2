package utils

import (
	"fmt"
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

	DBName     string `mapstructure:"DB_NAME"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBPort     string `mapstructure:"DB_PORT"`

	JwtExpiryInDays        int    `mapstructure:"JWT_EXPIRY_IN_DAYS"`
	JwtRefreshExpiryInDays int    `mapstructure:"JWT_REFRESH_EXPIRY_IN_DAYS"`
	JwtSecretKey           string `mapstructure:"JWT_SECRET_KEY"`

	Token string
	DBUrl string
	DBDsn string
}

func (c *Config) setDBString() {
	if c.DBPort == "" || c.DBPort == "nil" {
		c.DBUrl = fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=disable",
			c.DBUser,
			c.DBPassword,
			c.DBHost,
			c.DBName,
		)
		c.DBDsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
			c.DBHost,
			c.DBUser,
			c.DBPassword,
			c.DBName,
		)
	} else {
		c.DBUrl = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
			c.DBUser,
			c.DBPassword,
			c.DBHost,
			c.DBPort,
			c.DBName,
		)
		c.DBDsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			c.DBHost,
			c.DBUser,
			c.DBPassword,
			c.DBName,
			c.DBPort,
		)
	}
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
	config.setDBString()
	return
}
