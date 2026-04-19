package configs

import (
	"os"
)


var (
    secretKey = "JWT_SECRET_KEY"
	dsn = "DSN"
	serverAddr = "SERVER_ADDRESS"
)

type Config struct {
	SecretKey string
	DatabaseSoruceName string
	ServerAddress string
}


func GetConfig() *Config{
	config := &Config{
		SecretKey: "sXYKyBxAj+8mw3z58xeLV8AxBOevMA9eGajV+QOAwUA=",
		ServerAddress: "localhost:8080",
	}

	if value,ok:= os.LookupEnv(secretKey); ok {
		config.SecretKey = value
	}

	if value,ok:= os.LookupEnv(dsn); ok {
		config.DatabaseSoruceName = value
	}

	if value,ok:=os.LookupEnv(serverAddr); ok {
		config.ServerAddress = value
	}

	return config
}