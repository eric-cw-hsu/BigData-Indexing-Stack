package config

import (
	"log"
	"os"
	"path"

	"github.com/spf13/viper"
)

type Config struct {
	RabbitMQ struct {
		URI   string
		Queue string
	}
	ElasticSearch struct {
		Addr              string
		Index             string
		Username          string
		Password          string
		HealthCheckerPort string `mapstructure:"health_checker_port"`
	}
}

func Load() Config {
	dir, _ := os.Getwd()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath(path.Join(dir, "cmd/elasticsearch-service"))
	viper.AddConfigPath("config/")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unable to decode config into struct: %v", err)
	}

	return cfg
}
