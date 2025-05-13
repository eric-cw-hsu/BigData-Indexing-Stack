package main

import (
	"log"

	"eric-cw-hsu.github.io/internal/api/config"
	"eric-cw-hsu.github.io/internal/api/routes"
	"eric-cw-hsu.github.io/internal/database"
	"eric-cw-hsu.github.io/internal/rabbitmq"
)

func main() {
	cfg := config.Load()

	// Initialize MongoDB connection
	mongoService, err := database.NewMongoService(cfg.Mongo.URI, cfg.Mongo.Database)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoService.Close()
	planCollection := mongoService.GetCollection("plans")

	// Initialize RabbitMQ connection
	mq, err := rabbitmq.NewPublisher(cfg.RabbitMQ.URI, cfg.RabbitMQ.Queue)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer mq.Close()

	redisService := database.NewRedisService(cfg.Redis.URI)
	defer redisService.Close()
	redisClient := redisService.GetClient()

	router := routes.NewRouter(planCollection, mq, redisClient, cfg)
	router.Run(":" + cfg.Server.Port)
}
