package config

import "os"

type Config struct {
	DbHost     string
	DbPort     string
	DbUser     string
	DbPassword string
	DbName     string
}

func ParseConfig() Config {
	return Config{
		DbHost:     os.Getenv("POSTGRES_HOST"),
		DbPort:     os.Getenv("POSTGRES_PORT"),
		DbUser:     os.Getenv("POSTGRES_USER"),
		DbPassword: os.Getenv("POSTGRES_PASSWORD"),
		DbName:     os.Getenv("POSTGRES_DB"),
	}
}
