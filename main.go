package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/xeipuuv/gojsonschema"
	"golang.org/x/net/context"
)

var (
	rdb *redis.Client
	ctx = context.Background()

	jsonSchema = `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"planCostShares": {
				"type": "object",
				"properties": {
					"deductible": { "type": "integer" },
					"_org": { "type": "string" },
					"copay": { "type": "integer" },
					"objectId": { "type": "string" },
					"objectType": { "type": "string" }
				},
				"required": ["deductible", "_org", "copay", "objectId", "objectType"]
			},
			"linkedPlanServices": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"linkedService": {
							"type": "object",
							"properties": {
								"_org": { "type": "string" },
								"objectId": { "type": "string" },
								"objectType": { "type": "string" },
								"name": { "type": "string" }
							},
							"required": ["_org", "objectId", "objectType", "name"]
						},
						"planserviceCostShares": {
							"type": "object",
							"properties": {
								"deductible": { "type": "integer" },
								"_org": { "type": "string" },
								"copay": { "type": "integer" },
								"objectId": { "type": "string" },
								"objectType": { "type": "string" }
							},
							"required": ["deductible", "_org", "copay", "objectId", "objectType"]
						},
						"_org": { "type": "string" },
						"objectId": { "type": "string" },
						"objectType": { "type": "string" }
					},
					"required": ["linkedService", "planserviceCostShares", "_org", "objectId", "objectType"]
				}
			},
			"_org": { "type": "string" },
			"objectId": { "type": "string" },
			"objectType": { "type": "string" },
			"planType": { "type": "string" },
			"creationDate": { "type": "string"}
		},
		"required": ["planCostShares", "linkedPlanServices", "_org", "objectId", "objectType", "planType", "creationDate"]
	}`
)

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:16379",
		DB:   0,
	})

	r := gin.Default()

	r.POST("/v1/plan", createData)
	r.GET("/v1/plan/:key", readData)
	r.DELETE("/v1/plan/:key", deleteData)

	log.Println("Server running at :8080")
	r.Run(":8080")
}

// Generate ETag from payload
func generateETag(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%s", hash)
}

// Create or update data
func createData(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Validate JSON schema
	schemaLoader := gojsonschema.NewStringLoader(jsonSchema)
	documentLoader := gojsonschema.NewGoLoader(payload)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Schema validation error"})
		return
	}

	if !result.Valid() {
		errors := []string{}
		for _, err := range result.Errors() {
			errors = append(errors, err.String())
		}
		c.JSON(http.StatusBadRequest, gin.H{"validation_errors": errors})
		return
	}

	dataBytes, _ := json.Marshal(payload)
	etag := fmt.Sprintf("%x", generateETag(dataBytes))

	fmt.Println("dataBytes:", dataBytes)
	fmt.Println("String dataBytes:", string(dataBytes))

	// object id is used as key
	if err := rdb.Set(ctx, payload["objectId"].(string), dataBytes, 0).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store data"})
		return
	}

	fmt.Println("ETag:", etag)

	c.Header("ETag", etag)
	c.JSON(http.StatusCreated, gin.H{"message": "Data stored successfully"})
}

// Read data with conditional support
func readData(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'key' query parameter"})
		return
	}

	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read data"})
		return
	}

	fmt.Println("val:", val)

	dataBytes := []byte(val)
	etag := fmt.Sprintf("%x", generateETag(dataBytes))

	fmt.Println("ETag:", etag)
	fmt.Println("If-None-Match:", c.GetHeader("If-None-Match"))

	ifNoneMatch := c.GetHeader("If-None-Match")
	if ifNoneMatch == etag {
		c.Status(http.StatusNotModified)
		return
	}

	c.Header("ETag", etag)
	c.Data(http.StatusOK, "application/json", dataBytes)
}

// Delete data
func deleteData(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'key' query parameter"})
		return
	}

	if err := rdb.Del(ctx, key).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data"})
		return
	}

	c.Status(http.StatusNoContent)
}
