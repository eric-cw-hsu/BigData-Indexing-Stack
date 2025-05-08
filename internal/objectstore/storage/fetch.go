package storage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetNodeRaw(collection *mongo.Collection, id string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result map[string]interface{}
	if err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func GetExpandedNode(collection *mongo.Collection, id string) (map[string]interface{}, error) {
	rawNode, err := GetNodeRaw(collection, id)
	if err != nil {
		return nil, err
	}
	return expandRefs(collection, rawNode)
}

func expandRefs(collection *mongo.Collection, node map[string]interface{}) (map[string]interface{}, error) {
	for k, v := range node {
		switch vv := v.(type) {
		case map[string]interface{}:
			if ref, ok := vv["$ref"]; ok {
				referencedNode, err := GetExpandedNode(collection, ref.(string))
				if err != nil {
					return nil, err
				}
				node[k] = referencedNode
			}
		case []interface{}:
			expandedArray, err := expandArray(collection, vv)
			if err != nil {
				return nil, err
			}
			node[k] = expandedArray
		case primitive.A:
			expandedArray, err := expandArray(collection, []interface{}(vv))
			if err != nil {
				return nil, err
			}
			node[k] = expandedArray
		}
	}

	return node, nil
}

func expandArray(collection *mongo.Collection, items []interface{}) ([]interface{}, error) {
	expended := []interface{}{}
	for _, item := range items {
		refMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		refId, ok := refMap["$ref"].(string)
		if !ok {
			continue
		}
		referencedNode, err := GetExpandedNode(collection, refId)
		if err != nil {
			return nil, err
		}
		expended = append(expended, referencedNode)
	}

	return expended, nil
}
