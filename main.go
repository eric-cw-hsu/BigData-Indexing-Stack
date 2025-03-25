package main

import (
	"eric-cw-hsu.github.io/configs"
)

func main() {
	configs.SetupConfig()

	app := NewApp()
	app.RunAndServe(configs.AppConfig.Port)
}
