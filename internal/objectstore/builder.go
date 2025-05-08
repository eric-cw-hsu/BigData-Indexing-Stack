package objectstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func isNode(obj map[string]interface{}) bool {
	_, hasId := obj["objectId"]
	_, hasType := obj["objectType"]

	return hasId && hasType
}

func extractNodes(obj map[string]interface{}, nodes map[string]map[string]interface{}, parentId string, fieldName string) map[string]interface{} {
	if !isNode(obj) {
		return obj
	}

	objectId := obj["objectId"].(string)

	newNode := make(map[string]interface{})
	for k, v := range obj {
		switch t := v.(type) {
		case map[string]interface{}:
			val := extractNodes(t, nodes, objectId, k)
			if isNode(val) {
				refId := val["objectId"].(string)
				newNode[k] = map[string]interface{}{"$ref": refId}
				continue
			}
			newNode[k] = val
		case []interface{}:
			newArr := []interface{}{}
			for _, item := range t {
				if child, ok := item.(map[string]interface{}); ok {
					child = extractNodes(child, nodes, objectId, k)
					if isNode(child) {
						refId := child["objectId"].(string)
						newArr = append(newArr, map[string]interface{}{"$ref": refId})
					} else {
						newArr = append(newArr, child)
					}
				} else {
					newArr = append(newArr, item)
				}
			}
			newNode[k] = newArr
		default:
			newNode[k] = v
		}
	}

	newNode["parentId"] = parentId
	newNode["fieldName"] = fieldName
	nodes[objectId] = newNode
	return map[string]interface{}{"$ref": objectId}
}

func ExtractGraphNodes(index string, payload map[string]interface{}) map[string]map[string]interface{} {
	nodes := make(map[string]map[string]interface{})
	extractNodes(payload, nodes, "", index)
	return nodes
}

/*
*
  - StoreGraphNodes stores the nodes in the MongoDB collection.
  - It uses the objectId as the unique identifier for each node.
  - If a node with the same objectId already exists, update the refCount and other fields.
  - If the node is not found, it will be created.
  - If a node does not exist, it creates a new one.
*/
func StoreGraphNodes(collection *mongo.Collection, nodes map[string]map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for objectId, node := range nodes {
		node["_id"] = objectId

		update := bson.M{
			"$set": node,
		}

		if _, hasRefCount := node["refCount"]; !hasRefCount {
			update["$inc"] = bson.M{"refCount": 1}
		}

		opts := options.Update().SetUpsert(true)
		_, err := collection.UpdateOne(ctx, bson.M{"_id": objectId}, update, opts)
		if err != nil {
			return fmt.Errorf("failed to update node %s: %v", objectId, err)
		}
	}

	return nil
}

func getNode(collection *mongo.Collection, objectId string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetExpandedNode(collection *mongo.Collection, objectId string) (map[string]interface{}, error) {
	result, err := getNode(collection, objectId)
	if err != nil {
		return nil, err
	}

	return expandRefs(collection, result)
}

func expendArrayRefs(collection *mongo.Collection, array []interface{}) ([]interface{}, error) {
	expanded := []interface{}{}
	for _, item := range array {
		if refMap, ok := item.(map[string]interface{}); ok {
			if refId, ok := refMap["$ref"]; ok {
				refNode, err := GetExpandedNode(collection, refId.(string))
				if err != nil {
					return nil, err
				}

				expanded = append(expanded, refNode)
				continue
			}
		}
		expanded = append(expanded, item)
	}
	return expanded, nil
}

func expandRefs(collection *mongo.Collection, node map[string]interface{}) (map[string]interface{}, error) {
	for k, v := range node {
		switch val := v.(type) {
		case map[string]interface{}:
			if refId, ok := val["$ref"]; ok {
				referencedNode, err := GetExpandedNode(collection, refId.(string))
				if err != nil {
					return nil, err
				}
				node[k] = referencedNode
			}
		case []interface{}:
			expandedArray, err := expendArrayRefs(collection, val)
			if err != nil {
				return nil, err
			}
			node[k] = expandedArray
		case primitive.A:
			expandedArray, err := expendArrayRefs(collection, []interface{}(val))
			if err != nil {
				return nil, err
			}
			node[k] = expandedArray
		}
	}

	return node, nil
}

func DeleteGraphNodes(collection *mongo.Collection, nodes map[string]map[string]interface{}) error {
	for _, node := range nodes {
		err := deleteNode(collection, node)
		if err != nil {
			return fmt.Errorf("failed to delete node: %v", err)
		}
	}

	return nil
}

func deleteNode(collection *mongo.Collection, node map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectId := node["_id"].(string)
	refCount := node["refCount"].(int32)

	// update the refCount if there are other references on this node
	if refCount > 1 {
		update := bson.M{
			"$inc": bson.M{"refCount": -1},
		}
		_, err := collection.UpdateByID(ctx, objectId, update, options.Update().SetUpsert(false))
		if err != nil {
			return fmt.Errorf("failed to update refCount for node %s: %v", objectId, err)
		}

		return nil
	}

	// if no other references, delete the node
	_, err := collection.DeleteOne(ctx, bson.M{"_id": objectId})
	if err != nil {
		return fmt.Errorf("failed to delete node %s: %v", objectId, err)
	}

	return nil
}

func MergeJsonObjects(
	oriObject, updateObject map[string]interface{},
) (map[string]interface{}, map[string]map[string]interface{}, error) {
	toDeleteObjects := make(map[string]map[string]interface{})
	merged, err := mergeJsonObject(oriObject, updateObject, toDeleteObjects)
	if err != nil {
		return nil, nil, err
	}
	return merged, toDeleteObjects, nil
}

func mergeJsonObject(
	oriObject, updateObject map[string]interface{},
	toDeleteObjects map[string]map[string]interface{},
) (map[string]interface{}, error) {
	merged := make(map[string]interface{})
	oriId, oriHasId := oriObject["objectId"].(string)
	updateId, updateHasId := updateObject["objectId"].(string)

	// If both are nodes and have different objectIds → replace
	if oriHasId && updateHasId && oriId != updateId {
		toDeleteObjects[oriId] = oriObject
		return updateObject, nil
	}

	for key, oriVal := range oriObject {
		if updateVal, exists := updateObject[key]; exists {
			switch v := updateVal.(type) {
			case map[string]interface{}:
				mergedSub, err := mergeJsonObject(oriVal.(map[string]interface{}), v, toDeleteObjects)
				if err != nil {
					return nil, err
				}
				merged[key] = mergedSub

			case []interface{}:
				mergedArr, err := mergeArray(oriVal, v, toDeleteObjects)
				if err != nil {
					return nil, err
				}
				merged[key] = mergedArr

			default:
				merged[key] = updateVal
			}
		} else {
			merged[key] = oriVal
		}
	}

	// Append any new fields not in oriObject
	for key, updateVal := range updateObject {
		if _, exists := oriObject[key]; !exists {
			merged[key] = updateVal
		}
	}

	return merged, nil
}

func mergeArray(
	oriVal interface{},
	updateArr []interface{},
	toDeleteObjects map[string]map[string]interface{},
) ([]interface{}, error) {
	oriArray, ok := oriVal.([]interface{})
	if !ok {
		return nil, errors.New("original value is not an array")
	}

	oriMap := map[string]map[string]interface{}{}
	for _, item := range oriArray {
		if obj, ok := item.(map[string]interface{}); ok {
			if id, exists := obj["objectId"]; exists {
				oriMap[fmt.Sprintf("%v", id)] = obj
			}
		}
	}

	var result []interface{}
	for _, item := range updateArr {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		idStr := fmt.Sprintf("%v", obj["objectId"])
		if oriObj, found := oriMap[idStr]; found {
			merged, err := mergeJsonObject(oriObj, obj, toDeleteObjects)
			if err != nil {
				return nil, err
			}
			result = append(result, merged)
		} else {
			result = append(result, obj)
		}
	}

	return result, nil
}
