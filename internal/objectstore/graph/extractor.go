package graph

func isNode(obj map[string]interface{}) bool {
	_, hasId := obj["objectId"]
	_, hasType := obj["objectType"]

	return hasId && hasType
}

/*
ExtractGraphNodes extracts nodes from a given payload and returns a map of nodes.
*/
func ExtractGraphNodes(rootType string, payload map[string]interface{}) map[string]map[string]interface{} {
	nodes := make(map[string]map[string]interface{})
	extractNode(rootType, payload, "", nodes)
	return nodes
}

/*
extractNode recursively extracts nodes from the given object and its children.
*/
func extractNode(fieldName string, obj interface{}, parentId string, nodes map[string]map[string]interface{}) interface{} {
	m, ok := obj.(map[string]interface{})
	if !ok || !isNode(m) {
		return obj
	}

	id := m["objectId"].(string)
	node := make(map[string]interface{})
	for k, v := range m {
		switch vv := v.(type) {
		case map[string]interface{}:
			node[k] = extractNode(k, vv, id, nodes)
		case []interface{}:
			expandedArray := []interface{}{}
			for _, item := range vv {
				expandedArray = append(expandedArray, extractNode(k, item, id, nodes))
			}
			node[k] = expandedArray
		default:
			node[k] = vv
		}
	}

	node["parentId"] = parentId
	node["fieldName"] = fieldName
	nodes[id] = node

	return map[string]interface{}{"$ref": id}
}
