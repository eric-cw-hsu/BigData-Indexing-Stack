package config

import (
	"log"
	"os"
	"path"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port string
	}
	Mongo struct {
		URI      string
		Database string
	}
	OAuth struct {
		GoogleClientID string `mapstructure:"google_client_id"`
	}
	RabbitMQ struct {
		URI   string
		Queue string
	}
	Redis struct {
		URI string
	}
}

func Load() *Config {
	dir, _ := os.Getwd()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("config/")
	viper.AddConfigPath(path.Join(dir, "cmd/api-service"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	return &cfg
}
