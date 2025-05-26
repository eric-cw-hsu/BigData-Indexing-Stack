package storage

import (
	"context"
	"fmt"
	"time"

	"eric-cw-hsu.github.io/internal/shared/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func DeleteGraphNodes(collection *mongo.Collection, nodes map[string]map[string]interface{}) error {
	for _, node := range nodes {
		if err := deleteNode(collection, node); err != nil {
			logger.Logger.Error("storage.DeleteGraphNodes failed", zap.Error(err))
			return fmt.Errorf("failed to delete node: %v", err)
		}
	}
	return nil
}

func deleteNode(collection *mongo.Collection, node map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := node["_id"].(string)
	refCount := node["refCount"].(int32)
	if refCount > 1 {
		if _, err := collection.UpdateByID(ctx, id, map[string]interface{}{
			"$inc": map[string]interface{}{"refCount": -1},
		}); err != nil {
			logger.Logger.Error("storage.DeleteGraphNodes: failed to update refCount", zap.String("id", id), zap.Error(err))
			return fmt.Errorf("failed to update refCount for node %s: %v", id, err)
		}
		return nil
	}

	if _, err := collection.DeleteOne(ctx, map[string]interface{}{"_id": id}); err != nil {
		logger.Logger.Error("storage.DeleteGraphNodes: failed to delete node", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete node %s: %v", id, err)
	}
	return nil
}
