package configs

import (
	"errors"
	"log"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Environment string `mapstructure:"env"`
	Port        int    `mapstructure:"port"`
	LogLevel    logrus.Level
	Redis       RedisConfig  `mapstructure:"redis"`
	OAuth2      OAuth2Config `mapstructure:"oauth2"`
}

type RedisConfig struct {
	Addr string `mapstructure:"addr"`
	DB   int    `mapstructure:"db"`
}

func setRedisDefaultConfig() {
	viper.SetDefault("REDIS.ADDR", "localhost:6379")
	viper.SetDefault("REDIS.DB", 0)
}

type OAuth2Config struct {
	ClientID string `mapstructure:"client_id"`
	Issuer   string `mapstructure:"issuer"`
}

func (c *Config) Validate() error {
	if c.Port <= 0 {
		return errors.New("port must greater than 0")
	}

	if c.Environment == "" {
		return errors.New("environment must be set")
	}

	if c.Redis.Addr == "" {
		return errors.New("redis addr must be set")
	}

	if c.OAuth2.ClientID == "" {
		return errors.New("oauth2 client_id must be set")
	}

	if c.OAuth2.Issuer == "" {
		return errors.New("oauth2 issuer must be set")
	}

	return nil
}

var AppConfig Config

func SetupConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	setRedisDefaultConfig()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Unable to decode config into struct, %s", err)
	}

	if err := AppConfig.Validate(); err != nil {
		log.Fatalf("Invalid config, %s", err)
	}

	setLogLevel()
}

func setLogLevel() {
	switch AppConfig.Environment {
	case "development":
		AppConfig.LogLevel = logrus.DebugLevel
	case "production":
		AppConfig.LogLevel = logrus.InfoLevel
	case "testing":
		AppConfig.LogLevel = logrus.ErrorLevel
	default:
		AppConfig.LogLevel = logrus.InfoLevel
	}
}
