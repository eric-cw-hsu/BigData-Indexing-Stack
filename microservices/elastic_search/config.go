package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type ElasticSearchConfig struct {
	Addr              string `mapstructure:"addr"`
	Username          string `mapstructure:"username"`
	Password          string `mapstructure:"password"`
	HealthCheckerPort int    `mapstructure:"health_checker_port"`
}

type RabbitMQConfig struct {
	Addr string `mapstructure:"addr"`
}

type Config struct {
	ElasticSearch ElasticSearchConfig `mapstructure:"elasticsearch"`
	RabbitMQ      RabbitMQConfig      `mapstructure:"rabbitmq"`
}

func (c *Config) Validate() error {
	if c.ElasticSearch.Addr == "" {
		return errors.New("elastic_search addr must be set")
	}

	if c.ElasticSearch.Username == "" {
		return errors.New("elastic_search username must be set")
	}

	if c.ElasticSearch.Password == "" {
		return errors.New("elastic_search password must be set")
	}

	if c.RabbitMQ.Addr == "" {
		return errors.New("rabbitmq addr must be set")
	}

	if c.ElasticSearch.HealthCheckerPort == 0 {
		return errors.New("elastic_search health_checker_port must be set")
	}

	return nil
}

var AppConfig Config

func SetupConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Unable to decode config into struct, %s", err)
	}

	fmt.Println(AppConfig)

	if err := AppConfig.Validate(); err != nil {
		log.Fatalf("Invalid config, %s", err)
	}
}
