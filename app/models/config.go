package models

import "fmt"

type Config struct {
	RMApi string `mapstructure:"RM_API"`

	Port         string `mapstructure:"PORT"`
	PrintCsvLogs bool   `mapstructure:"PRINT_CSV_LOGS"`

	DBName     string `mapstructure:"DB_NAME"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBPort     string `mapstructure:"DB_PORT"`

	WsUpgradeWhitelist string `mapstructure:"WS_UPGRADE_WHITELIST"`
	WsUrl              string `mapstructure:"WS_URL"`

	// Coordinate interpolation settings
	EnableInterpolation     bool `mapstructure:"ENABLE_INTERPOLATION"`      // Enable/disable coordinate interpolation
	InterpolationIntervalMs int  `mapstructure:"INTERPOLATION_INTERVAL_MS"` // Interpolation interval in milliseconds

	JwtExpiryInDays        int    `mapstructure:"JWT_EXPIRY_IN_DAYS"`
	JwtRefreshExpiryInDays int    `mapstructure:"JWT_REFRESH_EXPIRY_IN_DAYS"`
	JwtSecretKey           string `mapstructure:"JWT_SECRET_KEY"`

	AdminApiKey string `mapstructure:"ADMIN_API_KEY"`

	Token string
	DBUrl string
	DBDsn string
}

func (c *Config) SetDBString() {
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

// SetInterpolationDefaults sets hardcoded default values for interpolation
func (c *Config) SetInterpolationDefaults() {
	// Hardcode interpolation to be enabled by default
	c.EnableInterpolation = true

	// Hardcode default interpolation interval to 1 second (1000ms)
	if c.InterpolationIntervalMs <= 0 {
		c.InterpolationIntervalMs = 1000
	}
}
