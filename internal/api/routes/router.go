package routes

import (
	"eric-cw-hsu.github.io/internal/api/config"
	"eric-cw-hsu.github.io/internal/api/handlers"
	"eric-cw-hsu.github.io/internal/api/repositories"
	"eric-cw-hsu.github.io/internal/api/services"
	"eric-cw-hsu.github.io/internal/shared/logger"
	"eric-cw-hsu.github.io/internal/shared/messagequeue"
	"eric-cw-hsu.github.io/internal/shared/middleware"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewRouter(collection *mongo.Collection, publisher *messagequeue.Publisher, redisClient *redis.Client, config *config.Config) *gin.Engine {
	planRepository := repositories.NewPlanRepository(collection)
	planService := services.NewPlanService(publisher, planRepository, redisClient)
	planHandler := handlers.NewPlanHandler(planRepository, planService)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(middleware.RecoveryWithLogger(logger.Logger))
	router.Use(middleware.PrometheusMiddleware())

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// router.Use(oauth.GoogleAuthMiddleware(config.OAuth.GoogleClientID))

	router.GET("v1/plans/:id", planHandler.GetPlanHandler)
	router.POST("/v1/plans", planHandler.StorePlanHandler)
	router.DELETE("/v1/plans/:id", planHandler.DeletePlanHandler)
	router.PATCH("/v1/plans/:id", planHandler.UpdatePlanHandler)

	return router
}
