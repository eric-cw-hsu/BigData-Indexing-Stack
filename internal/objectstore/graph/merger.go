package graph

func Merge(
	original, update map[string]interface{},
) (merged map[string]interface{}, toDelete map[string]map[string]interface{}, err error) {
	toDelete = make(map[string]map[string]interface{})
	merged, err = mergeObject(original, update, toDelete)
	return
}

func mergeObject(
	original, update map[string]interface{},
	toDelete map[string]map[string]interface{},
) (map[string]interface{}, error) {
	merged := make(map[string]interface{})
	oriId, oriHasId := original["objectId"].(string)
	updateId, updateHasId := update["objectId"].(string)

	// If both are nodes and have different objectIds -> replace
	if oriHasId && updateHasId && oriId != updateId {
		toDelete[oriId] = original
		return update, nil
	}

	for key, oriVal := range original {
		if updateVal, exists := update[key]; exists {
			switch v := updateVal.(type) {
			case map[string]interface{}:
				mergedSub, err := mergeObject(oriVal.(map[string]interface{}), v, toDelete)
				if err != nil {
					return nil, err
				}
				merged[key] = mergedSub
			case []interface{}:
				mergedArr, err := mergeArray(oriVal.([]interface{}), v, toDelete)
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

	for key, updateVal := range update {
		if _, exists := original[key]; !exists {
			merged[key] = updateVal
		}
	}
	return merged, nil
}

/*
MergeArray merges two arrays of objects. Update objects that exist in both arrays and merge rest of the objects.
*/
func mergeArray(
	originalArray, updateArray []interface{},
	toDelete map[string]map[string]interface{},
) ([]interface{}, error) {
	// Create a map for the original array with [objectId] as keys
	originalMap := map[string]map[string]interface{}{}
	for _, item := range originalArray {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if id, exists := itemMap["objectId"]; exists {
				originalMap[id.(string)] = itemMap
			}
		}
	}

	mergedArray := []interface{}{}
	for _, item := range updateArray {
		updateItemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		if !isNode(updateItemMap) {
			continue
		}

		id := updateItemMap["objectId"].(string)
		originalItem, exists := originalMap[id]

		// If the item does not exist in originalArray, add it to mergedArray
		if !exists {
			mergedArray = append(mergedArray, item)
			continue
		}

		// If the item exists in originalArray, merge it
		mergedItem, err := mergeObject(originalItem, updateItemMap, toDelete)
		if err != nil {
			return nil, err
		}
		mergedArray = append(mergedArray, mergedItem)
		delete(originalMap, id)
	}

	for _, item := range originalMap {
		mergedArray = append(mergedArray, item)
	}

	return mergedArray, nil
}
