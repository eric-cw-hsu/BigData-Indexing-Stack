package controllers

import (
	"net/http"

	"eric-cw-hsu.github.io/errors"
	"eric-cw-hsu.github.io/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type PlanController struct {
	planService *services.PlanService
	logger      *logrus.Logger
}

func NewPlanController(planService *services.PlanService, logger *logrus.Logger) *PlanController {
	return &PlanController{
		planService: planService,
		logger:      logger,
	}
}

func (c *PlanController) CreatePlan() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var payload map[string]interface{}
		if err := ctx.ShouldBindJSON(&payload); err != nil {
			c.logger.Error(errors.ErrInvalidJSON, err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidJSON.Error()})
			return
		}

		key, etag, err := c.planService.CreatePlan(payload)
		if err != nil {
			if err == errors.ErrKeyAlreadyExists {
				ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			} else {
				ctx.Error(err)
			}
			return
		}

		ctx.Header("ETag", etag)
		ctx.JSON(http.StatusCreated, gin.H{"message": "Data stored successfully", "key": key})
	}
}

func (c *PlanController) GetPlan() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := ctx.Param("key")
		if key == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrMissingKeyQueryParam.Error()})
			return
		}

		val, etag, err := c.planService.GetPlan(key)
		if err != nil {
			if err == errors.ErrKeyNotFound {
				ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			} else {
				ctx.Error(err)
			}
			return
		}

		if ctx.GetHeader("If-None-Match") == etag {
			ctx.Status(http.StatusNotModified)
			return
		}

		ctx.Header("ETag", etag)
		ctx.Data(http.StatusOK, "application/json", []byte(val))
	}
}

func (c *PlanController) DeletePlan() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := ctx.Param("key")
		if key == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrMissingKeyQueryParam.Error()})
			return
		}

		if err := c.planService.DeletePlan(key); err != nil {
			if err == errors.ErrKeyNotFound {
				ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			} else {
				ctx.Error(err)
			}
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Data deleted successfully"})
	}
}

func (c *PlanController) UpdatePlan() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := ctx.Param("key")
		if key == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrMissingKeyQueryParam.Error()})
			return
		}

		var payload map[string]interface{}
		if err := ctx.ShouldBindJSON(&payload); err != nil {
			c.logger.Error(errors.ErrInvalidJSON, err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidJSON.Error()})
			return
		}

		val, etag, err := c.planService.GetPlan(key)
		if err != nil {
			if err == errors.ErrKeyNotFound {
				ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			} else {
				ctx.Error(err)
			}
			return
		}

		if ctx.GetHeader("If-None-Match") == "" {
			ctx.JSON(http.StatusPreconditionRequired, gin.H{"error": errors.ErrMissingIfNoneMatch.Error()})
			return
		}

		if ctx.GetHeader("If-None-Match") != etag {
			ctx.JSON(http.StatusPreconditionFailed, gin.H{"error": errors.ErrPreconditionFailed.Error()})
			return
		}

		val, etag, err = c.planService.UpdatePlan(key, payload)
		if err != nil {
			if err == errors.ErrKeyNotFound {
				ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			} else {
				ctx.Error(err)
			}
			return
		}

		ctx.Header("ETag", etag)
		ctx.Data(http.StatusOK, "application/json", []byte(val))
	}
}
