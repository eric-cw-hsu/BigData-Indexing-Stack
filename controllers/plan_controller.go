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

		err := utils.ValidateJSONSchema(payload, models.GetPlanJsonSchema())
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		key := payload["objectType"].(string) + ":" + payload["objectId"].(string)

		// check if the key already exists
		if _, err := c.planRepository.GetPlan(key); err != redis.Nil {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Key already exists"})
			return
		}

		dataBytes, _ := json.Marshal(payload)

		if err := c.planRepository.StorePlan(key, dataBytes); err != nil {
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
		key := "plan:" + ctx.Param("key")
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
		key := "plan:" + ctx.Param("key")
		if key == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'key' query parameter"})
			return
		}

		if _, err := c.planRepository.GetPlan(key); err == redis.Nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
			return
		}

		if err := c.planRepository.DeletePlan(key); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data"})
			return
		}

		ctx.Status(http.StatusNoContent)
	}
}

func (c *PlanController) UpdatePlan() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := "plan:" + ctx.Param("key")
		if key == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'key' query parameter"})
			return
		}

		var payload map[string]interface{}
		if err := ctx.ShouldBindJSON(&payload); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		val, err := c.planRepository.GetPlan(key)
		if err == redis.Nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
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
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse existing data"})
			return
		}

		// 合併數據：覆蓋頂層屬性，合併 linkedPlanServices
		mergedData := mergePlans(existingData, payload)

		// 驗證 JSON Schema
		if err := utils.ValidateJSONSchema(mergedData, models.GetPlanJsonSchema()); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		dataBytes, _ := json.Marshal(mergedData)
		if err := c.planRepository.StorePlan(key, dataBytes); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update data"})
			return
		}

		etag = fmt.Sprintf("%x", utils.GenerateETag(dataBytes))
		ctx.Header("ETag", etag)
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
