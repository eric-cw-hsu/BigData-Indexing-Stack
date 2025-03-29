package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

func main() {
	// Setup configuration and logger.
	SetupConfig()
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Initialize RabbitMQQueue.
	rabbitMQConn, err := amqp091.Dial(AppConfig.RabbitMQ.Addr)
	if err != nil {
		logger.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	rabbitChannel, err := rabbitMQConn.Channel()
	if err != nil {
		logger.Fatalf("Failed to open a channel: %v", err)
	}

	// Initialize ElasticSearchService.
	elasticService := NewElasticSearchService(logger)

	// Create ElasticConsumer using the RabbitMQ channel.
	elasticConsumer := NewElasticConsumer(rabbitChannel, "elastic", elasticService, logger)

	// Start consumer in background.
	go func() {
		if err := elasticConsumer.StartConsuming(context.Background()); err != nil {
			logger.Errorf("ElasticConsumer encountered an error: %v", err)
		}
	}()

	// Start HTTP health checker.
	startHTTPHealthChecker(logger)
}

func startHTTPHealthChecker(logger *logrus.Logger) {
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	addr := fmt.Sprintf(":%d", AppConfig.ElasticSearch.HealthCheckerPort)
	logger.Infof("Elastic Consumer Microservice starting on port %d", AppConfig.ElasticSearch.HealthCheckerPort)
	if err := router.Run(addr); err != nil {
		logger.Fatalf("HTTP server error: %v", err)
	}
}
