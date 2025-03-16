package routes

import (
	"eric-cw-hsu.github.io/controllers"
	"eric-cw-hsu.github.io/middlewares"
	"eric-cw-hsu.github.io/repositories"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func PlanRoute(router *gin.RouterGroup, redis *redis.Client) {
	planRepository := repositories.NewPlanRepository(redis)
	planController := controllers.NewPlanController(planRepository)

	planRouter := router.Group("/plan")
	planRouter.Use(middlewares.AuthenticateHandler())
	{
		planRouter.POST("/", planController.CreatePlan())
		planRouter.GET("/:key", planController.GetPlan())
		planRouter.PATCH("/:key", planController.UpdatePlan())
		planRouter.DELETE("/:key", planController.DeletePlan())
	}
}
