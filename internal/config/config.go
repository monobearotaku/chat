package config

import "os"

type Db struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type Kafka struct {
	Broker string
	Topic  string
}

type Tracer struct {
	Url string
}

type Config struct {
	Db     Db
	Kafka  Kafka
	Tracer Tracer
}

func ParseConfig() Config {
	return Config{
		Db: Db{
			Host:     os.Getenv("POSTGRES_HOST"),
			Port:     os.Getenv("POSTGRES_PORT"),
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			Name:     os.Getenv("POSTGRES_DB"),
		},
		Kafka: Kafka{
			Broker: os.Getenv("KAFKA_BROKER"),
			Topic:  os.Getenv("KAFKA_TOPIC"),
		},
		Tracer: Tracer{
			Url: os.Getenv("TRACER_URL"),
		},
	}
}
