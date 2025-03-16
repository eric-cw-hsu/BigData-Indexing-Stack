package main

import (
	"eric-cw-hsu.github.io/configs"
	"github.com/spf13/viper"
)

func setupConfig() {
	configs.SetDefaultConfig()
	configs.SetRedisDefaultConfig()
	configs.SetOAuth2DefaultConfig()

	configs.GetEnvironment()
}

func main() {
	setupConfig()

	app := NewApp()
	app.RunAndServe(viper.GetInt("PORT"))
}
