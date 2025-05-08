package storage

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func StoreExtractedGraphNodes(collection *mongo.Collection, nodes map[string]map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for id, node := range nodes {
		node["_id"] = id
		update := bson.M{"$set": node}
		if _, exists := node["refCount"]; !exists {
			update["$inc"] = bson.M{"refCount": 1}
		}
		opts := options.Update().SetUpsert(true)
		if _, err := collection.UpdateByID(ctx, id, update, opts); err != nil {
			return fmt.Errorf("failed to store node %s: %v", id, err)
		}
	}

	return nil
}
