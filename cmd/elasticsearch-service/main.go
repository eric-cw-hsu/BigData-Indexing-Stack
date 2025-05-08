package main

import (
	"fmt"
	"net/http"

	"eric-cw-hsu.github.io/internal/elasticsearch"
	"eric-cw-hsu.github.io/internal/elasticsearch/config"
	"eric-cw-hsu.github.io/internal/rabbitmq"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// Initialize RabbitMQ Consumer
	rabbitMQConsumer, err := rabbitmq.NewConsumer(cfg.RabbitMQ.URI, cfg.RabbitMQ.Queue)
	if err != nil {
		panic(fmt.Sprintf("Failed to create RabbitMQ consumer: %v", err))
	}

	// Initialize ElasticSearch Client
	esClient, err := elasticsearch.NewElasticSearchClient(
		cfg.ElasticSearch.Addr,
		cfg.ElasticSearch.Username,
		cfg.ElasticSearch.Password,
		cfg.ElasticSearch.Index,
	)

	elasticsearch.Start(rabbitMQConsumer, esClient)

	// Start the health check server
	fmt.Println("Starting health check server...")
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.Run(fmt.Sprintf(":%s", cfg.ElasticSearch.HealthCheckerPort))
}
