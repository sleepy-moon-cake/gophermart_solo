package configs

import (
	"flag"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	SecretKey            string
	DatabaseSoruceName   string
	ServerAddress        string
	AccrualSystemAddress string
}

func GetConfig() *Config {
	a := flag.String("a", "localhost:8080", "server address")
	d := flag.String("d", "", "database uri")
	r := flag.String("r", "http://localhost:8081", "accrual system address")
	flag.Parse()

	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		slog.Error("JWT_SECRET_KEY environment variable is required")
		os.Exit(1)
	}

	config := &Config{
		SecretKey:            "sXYKyBxAj+8mw3z58xeLV8AxBOevMA9eGajV+QOAwUA=",
		ServerAddress:        *a,
		DatabaseSoruceName:   *d,
		AccrualSystemAddress: *r,
	}

	if val := os.Getenv("RUN_ADDRESS"); val != "" {
		config.ServerAddress = val
	}

	if val := os.Getenv("DATABASE_URI"); val != "" {
		config.DatabaseSoruceName = val
	}

	if val := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); val != "" {
		config.AccrualSystemAddress = val
	}

	if val := os.Getenv("JWT_SECRET_KEY"); val != "" {
		config.SecretKey = val
	}

	if config.AccrualSystemAddress != "" && !strings.HasPrefix(config.AccrualSystemAddress, "http") {
		config.AccrualSystemAddress = "http://" + config.AccrualSystemAddress
	}

	return config
}
