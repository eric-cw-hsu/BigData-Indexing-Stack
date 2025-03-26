package routes

import (
	"eric-cw-hsu.github.io/controllers"
	"eric-cw-hsu.github.io/middlewares"
	"eric-cw-hsu.github.io/repositories"
	"eric-cw-hsu.github.io/services"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

func PlanRoute(router *gin.RouterGroup, redis *redis.Client, logger *logrus.Logger) {
	planRepository := repositories.NewPlanRepository(redis, logger)
	planService := services.NewPlanService(planRepository, logger)
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
