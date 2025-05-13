package routes

import (
	"eric-cw-hsu.github.io/internal/api/config"
	"eric-cw-hsu.github.io/internal/api/handlers"
	"eric-cw-hsu.github.io/internal/api/repositories"
	"eric-cw-hsu.github.io/internal/api/services"
	"eric-cw-hsu.github.io/internal/oauth"
	"eric-cw-hsu.github.io/internal/rabbitmq"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewRouter(collection *mongo.Collection, publisher *rabbitmq.Publisher, redisClient *redis.Client, config *config.Config) *gin.Engine {
	planRepository := repositories.NewPlanRepository(collection)
	planService := services.NewPlanService(publisher, planRepository, redisClient)
	planHandler := handlers.NewPlanHandler(planRepository, planService)

	router := gin.Default()

	router.Use(oauth.GoogleAuthMiddleware(config.OAuth.GoogleClientID))

	router.GET("v1/plans/:id", planHandler.GetPlanHandler)
	router.POST("/v1/plans", planHandler.StorePlanHandler)
	router.DELETE("/v1/plans/:id", planHandler.DeletePlanHandler)
	router.PATCH("/v1/plans/:id", planHandler.UpdatePlanHandler)

	return router
}
