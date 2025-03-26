package routes

import (
	"eric-cw-hsu.github.io/controllers"
	"eric-cw-hsu.github.io/middlewares"
	"eric-cw-hsu.github.io/repositories"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

func PlanRoute(router *gin.RouterGroup, redis *redis.Client, logger *logrus.Logger) {
	planRepository := repositories.NewPlanRepository(redis, logger)
	planController := controllers.NewPlanController(planRepository, logger)

	planRouter := router.Group("/plan")
	planRouter.Use(middlewares.AuthenticateHandler())
	{
		planRouter.POST("/", planController.CreatePlan())
		planRouter.GET("/:key", planController.GetPlan())
		planRouter.PATCH("/:key", planController.UpdatePlan())
		planRouter.DELETE("/:key", planController.DeletePlan())
	}
}
