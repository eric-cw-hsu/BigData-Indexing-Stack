package main

import (
	"fmt"

	"eric-cw-hsu.github.io/configs"
	"eric-cw-hsu.github.io/middlewares"
	"eric-cw-hsu.github.io/routes"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type App struct {
	router *gin.Engine
	redis  *redis.Client
	logger *logrus.Logger
}

func NewApp() *App {
	configs.SetupConfig()

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(configs.AppConfig.LogLevel)

	app := &App{
		router: gin.New(),
		redis: redis.NewClient(&redis.Options{
			Addr: configs.AppConfig.Redis.Addr,
			DB:   configs.AppConfig.Redis.DB,
		}),
		logger: logger,
	}

	app.setupMiddlewares()
	app.setupRoutes()

	return app
}

func (app *App) setupMiddlewares() {
	app.router.Use(gin.Recovery())
	app.router.Use(middlewares.ErrorHandler(app.logger))
	app.router.Use(middlewares.RequestLogger(app.logger))
}

func (app *App) setupRoutes() {
	v1Router := app.router.Group("/v1")
	{
		routes.PlanRoute(v1Router, app.redis, app.logger)
	}
}

func (app *App) RunAndServe(port int) {
	addr := fmt.Sprintf(":%d", port)
	app.logger.Infof("Starting server on port %d", port)
	if err := app.router.Run(addr); err != nil {
		app.logger.Fatalf("Failed to start server: %v", err)
	}
}
