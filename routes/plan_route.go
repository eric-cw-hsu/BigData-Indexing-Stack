package routes

import (
	"eric-cw-hsu.github.io/configs"
	"eric-cw-hsu.github.io/controllers"
	"eric-cw-hsu.github.io/middlewares"
	"eric-cw-hsu.github.io/queue"
	"eric-cw-hsu.github.io/repositories"
	"eric-cw-hsu.github.io/services"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

func PlanRoute(router *gin.RouterGroup, redisClient *redis.Client, logger *logrus.Logger) {
	rabbitQueue, err := queue.NewRabbitMQQueue(configs.AppConfig.RabbitMQ.Addr, "elastic", logger)
	if err != nil {
		logger.Error("Failed to initialize RabbitMQQueue, proceeding without queue: ", err)
		rabbitQueue = nil
	}

	planRepository := repositories.NewPlanRepository(redisClient, logger)
	planService := services.NewPlanService(planRepository, logger, rabbitQueue)
	planController := controllers.NewPlanController(planService, logger)

	planRouter := router.Group("/plan")
	planRouter.Use(middlewares.AuthenticateHandler())
	{
		planRouter.POST("/", planController.CreatePlan())
		planRouter.GET("/:key", planController.GetPlan())
		planRouter.PATCH("/:key", planController.UpdatePlan())
		planRouter.DELETE("/:key", planController.DeletePlan())
	}
}
