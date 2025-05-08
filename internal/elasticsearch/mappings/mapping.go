package mappings

func GetPlanMapping() string {
	return `{
			"mappings": {
				"properties": {
					"planCostShares": {
						"type": "object",
						"properties": {
							"deductible": { "type": "integer" },
							"_org": { "type": "keyword" },
							"copay": { "type": "integer" },
							"objectId": { "type": "keyword" },
							"objectType": { "type": "keyword" }
						}
					},
					"linkedPlanServices": {
						"type": "nested",
						"properties": {
							"linkedService": {
								"type": "object",
								"properties": {
									"_org": { "type": "keyword" },
									"objectId": { "type": "keyword" },
									"objectType": { "type": "keyword" },
									"name": { "type": "text" }
								}
							},
							"planserviceCostShares": {
								"type": "object",
								"properties": {
									"deductible": { "type": "integer" },
									"_org": { "type": "keyword" },
									"copay": { "type": "integer" },
									"objectId": { "type": "keyword" },
									"objectType": { "type": "keyword" }
								}
							},
							"_org": { "type": "keyword" },
							"objectId": { "type": "keyword" },
							"objectType": { "type": "keyword" }
						}
					},
					"_org": { "type": "keyword" },
					"objectId": { "type": "keyword" },
					"objectType": { "type": "keyword" },
					"planType": { "type": "keyword" },
					"creationDate": { "type": "text" },
					"join_field": {
						"type": "join",
						"relations": {
							"plan": ["planCostShares", "linkedPlanServices"],
							"linkedPlanServices": ["linkedService", "planserviceCostShares"]
						}
					}
				}
			}
		}`
}
