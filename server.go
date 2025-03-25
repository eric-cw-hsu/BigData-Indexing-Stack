package main

import (
	"fmt"

	"eric-cw-hsu.github.io/configs"
	"eric-cw-hsu.github.io/routes"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type App struct {
	router *gin.Engine
	redis  *redis.Client
}

func NewApp() *App {
	app := &App{
		redis: redis.NewClient(&redis.Options{
			Addr: configs.AppConfig.Redis.Addr,
			DB:   configs.AppConfig.Redis.DB,
		}),
	}

	app.setupRoutes()

	return app
}

func (app *App) setupRoutes() {
	app.router = gin.Default()

	v1Router := app.router.Group("/v1")
	{
		routes.PlanRoute(v1Router, app.redis)
	}
}

func (app *App) RunAndServe(port int) {
	app.router.Run(fmt.Sprint(":", port))
}
