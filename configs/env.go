package configs

import (
	"log"

	"github.com/spf13/viper"
)

func GetEnvironment() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file, %s", err)
	}
}

func SetDefaultConfig() {
	viper.SetDefault("PORT", 8080)
}
