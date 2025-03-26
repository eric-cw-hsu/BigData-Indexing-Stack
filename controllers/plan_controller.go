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
	"github.com/sirupsen/logrus"
)

type PlanController struct {
	planRepository repositories.IPlanRepository
	logger         *logrus.Logger
}

func NewPlanController(planRepository repositories.IPlanRepository, logger *logrus.Logger) *PlanController {
	return &PlanController{
		planRepository: planRepository,
		logger:         logger,
	}
}

func (c *PlanController) CreatePlan() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var payload map[string]interface{}
		if err := ctx.ShouldBindJSON(&payload); err != nil {
			c.logger.Error("Invalid JSON payload:", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		err := utils.ValidateJSONSchema(payload, models.GetPlanJsonSchema())
		if err != nil {
			c.logger.Error("JSON Schema validation error:", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		key := payload["objectType"].(string) + ":" + payload["objectId"].(string)

		if _, err := c.planRepository.GetPlan(key); err != redis.Nil {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Key already exists"})
			return
		}

		dataBytes, err := json.Marshal(payload)
		if err != nil {
			c.logger.Error("JSON Marshal error:", err)
			ctx.Error(err)
			return
		}

		if err := c.planRepository.StorePlan(key, dataBytes); err != nil {
			c.logger.Error("StorePlan error:", err)
			ctx.Error(err)
			return
		}

		etag := fmt.Sprintf("%x", utils.GenerateETag(dataBytes))
		ctx.Header("ETag", etag)
		c.logger.Infof("Plan created: %s", key)
		ctx.JSON(http.StatusCreated, gin.H{"message": "Data stored successfully"})
	}
}

func (c *PlanController) GetPlan() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := "plan:" + ctx.Param("key")
		if key == "plan:" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'key' query parameter"})
			return
		}

		val, err := c.planRepository.GetPlan(key)
		if err == redis.Nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
			return
		} else if err != nil {
			c.logger.Error("GetPlan error:", err)
			ctx.Error(err)
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
		key := "plan:" + ctx.Param("key")
		if key == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'key' query parameter"})
			return
		}

		if _, err := c.planRepository.GetPlan(key); err == redis.Nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
			return
		} else if err != nil {
			c.logger.Error("GetPlan error in DeletePlan:", err)
			ctx.Error(err)
			return
		}

		if err := c.planRepository.DeletePlan(key); err != nil {
			c.logger.Error("DeletePlan error:", err)
			ctx.Error(err)
			return
		}

		c.logger.Infof("Plan deleted: %s", key)
		ctx.Status(http.StatusNoContent)
	}
}

func (c *PlanController) UpdatePlan() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := "plan:" + ctx.Param("key")
		if key == "plan:" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'key' query parameter"})
			return
		}

		var payload map[string]interface{}
		if err := ctx.ShouldBindJSON(&payload); err != nil {
			c.logger.Error("Invalid JSON in UpdatePlan:", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		val, err := c.planRepository.GetPlan(key)
		if err == redis.Nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
			return
		} else if err != nil {
			c.logger.Error("GetPlan error in UpdatePlan:", err)
			ctx.Error(err)
			return
		}

		etag := fmt.Sprintf("%x", utils.GenerateETag([]byte(val)))
		if ctx.GetHeader("If-None-Match") == "" {
			ctx.JSON(http.StatusPreconditionRequired, gin.H{"error": "Missing 'If-None-Match' header"})
			return
		}

		if ctx.GetHeader("If-None-Match") != etag {
			ctx.JSON(http.StatusPreconditionFailed, gin.H{"error": "Precondition failed"})
			return
		}

		existingData := make(map[string]interface{})
		if err := json.Unmarshal([]byte(val), &existingData); err != nil {
			c.logger.Error("Unmarshal error in UpdatePlan:", err)
			ctx.Error(err)
			return
		}

		mergedData := mergePlans(existingData, payload)

		if err := utils.ValidateJSONSchema(mergedData, models.GetPlanJsonSchema()); err != nil {
			c.logger.Error("JSON Schema validation error:", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		dataBytes, err := json.Marshal(mergedData)
		if err != nil {
			c.logger.Error("JSON Marshal error in UpdatePlan:", err)
			ctx.Error(err)
			return
		}

		if err := c.planRepository.StorePlan(key, dataBytes); err != nil {
			c.logger.Error("StorePlan error in UpdatePlan:", err)
			ctx.Error(err)
			return
		}

		etag = fmt.Sprintf("%x", utils.GenerateETag(dataBytes))
		ctx.Header("ETag", etag)
		c.logger.Infof("Plan updated: %s", key)
		ctx.Data(http.StatusOK, "application/json", dataBytes)
	}
}

func mergePlans(existing, newPlan map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range existing {
		merged[k] = v
	}

	for k, v := range newPlan {
		if k == "linkedPlanServices" {
			merged[k] = mergeLinkedPlanServices(existing[k], v)
		} else {
			merged[k] = v
		}
	}

	return merged
}

func mergeLinkedPlanServices(existing, newData interface{}) []map[string]interface{} {
	existingList, ok1 := existing.([]interface{})
	newList, ok2 := newData.([]interface{})
	if !ok1 {
		existingList = []interface{}{}
	}
	if !ok2 {
		newList = []interface{}{}
	}

	existingMap := make(map[string]map[string]interface{})
	for _, item := range existingList {
		if obj, ok := item.(map[string]interface{}); ok {
			if id, exists := obj["objectId"].(string); exists {
				existingMap[id] = obj
			}
		}
	}

	for _, item := range newList {
		if obj, ok := item.(map[string]interface{}); ok {
			if id, exists := obj["objectId"].(string); exists {
				existingMap[id] = obj
			}
		}
	}

	mergedList := []map[string]interface{}{}
	for _, obj := range existingMap {
		mergedList = append(mergedList, obj)
	}

	return mergedList
}
