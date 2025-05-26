package main

import (
	"log"

	"eric-cw-hsu.github.io/internal/api/config"
	"eric-cw-hsu.github.io/internal/api/routes"
	"eric-cw-hsu.github.io/internal/database"
	"eric-cw-hsu.github.io/internal/rabbitmq"
	"eric-cw-hsu.github.io/internal/shared/logger"
	"eric-cw-hsu.github.io/internal/shared/messagequeue"
	"go.uber.org/zap"
)

func main() {
	if err := logger.InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	cfg := config.Load()

	// Initialize MongoDB connection
	mongoService, err := database.NewMongoService(cfg.Mongo.URI, cfg.Mongo.Database)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoService.Close()
	planCollection := mongoService.GetCollection("plans")

	// Initialize RabbitMQ connection
	rabbitmqConn, err := rabbitmq.NewMQConnection(cfg.RabbitMQ.URI)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer rabbitmqConn.Close()
	publisher, err := messagequeue.NewPublisher(rabbitmqConn.Channel, cfg.RabbitMQ.Exchange)

	redisService := database.NewRedisService(cfg.Redis.URI)
	defer redisService.Close()
	redisClient := redisService.GetClient()

	router := routes.NewRouter(planCollection, publisher, redisClient, cfg)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		logger.Logger.Fatal("Failed to start server", zap.Error(err))
	}
}
