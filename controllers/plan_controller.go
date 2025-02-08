package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"eric-cw-hsu.github.io/models"
	"eric-cw-hsu.github.io/repositories"
	"eric-cw-hsu.github.io/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/xeipuuv/gojsonschema"
)

type PlanController struct {
	planRepository repositories.IPlanRepository
}

func NewPlanController(planRepository repositories.IPlanRepository) *PlanController {
	return &PlanController{
		planRepository: planRepository,
	}
}

func (c *PlanController) CreatePlan() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var payload map[string]interface{}
		if err := ctx.ShouldBindJSON(&payload); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid JSON",
			})
			return
		}

		// validate JSON Schema
		schemaLoader := gojsonschema.NewStringLoader(models.GetPlanJsonSchema())
		documentLoader := gojsonschema.NewGoLoader(payload)

		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Schema validation error",
			})
		}

		if !result.Valid() {
			errors := []string{}
			for _, err := range result.Errors() {
				errors = append(errors, err.String())
			}
			ctx.JSON(http.StatusBadRequest, gin.H{"validation_errors": errors})
			return
		}

		dataBytes, _ := json.Marshal(payload)

		if err := c.planRepository.StorePlan(payload["objectId"].(string), dataBytes); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store data"})
			return
		}

		etag := fmt.Sprintf("%x", utils.GenerateETag(dataBytes))
		ctx.Header("ETag", etag)
		ctx.JSON(http.StatusCreated, gin.H{
			"message": "Data stored successfully",
		})
	}
}

func (c *PlanController) GetPlan() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := ctx.Param("key")
		if key == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'key' query parameter"})
			return
		}

		val, err := c.planRepository.GetPlan(key)
		if err == redis.Nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
			return
		} else if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read data"})
			return
		}

		dataBytes := []byte(val)
		etag := fmt.Sprintf("%x", utils.GenerateETag(dataBytes))

		if ctx.GetHeader("If-None-Match") == etag {
			ctx.Status(http.StatusNotModified)
			return
		}

		ctx.Header("ETag", etag)
		ctx.Data(http.StatusOK, "application/json", dataBytes)
	}
}

func (c *PlanController) DeletePlan() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := ctx.Param("key")
		if key == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'key' query parameter"})
			return
		}

		if err := c.planRepository.DeletePlan(key); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data"})
			return
		}

		ctx.Status(http.StatusNoContent)
	}
}
