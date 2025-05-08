package elasticsearch

import (
	"encoding/json"
	"log"
	"strings"

	"eric-cw-hsu.github.io/internal/rabbitmq/types"
)

func ProcessRabbitMQPlanMessage(client *Client, body []byte) {
	var msg types.PlanMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("Invalid message format: %v", err)
		return
	}

	parentId, ok := msg.Data["parentId"].(string)
	if !ok {
		parentId = ""
	}

	switch strings.ToLower(msg.Action) {
	case "create", "update":
		client.IndexDocument(msg.Key, msg.Data, parentId)
	case "delete":
		client.DeleteDocument(msg.Key)
	default:
		log.Printf("Unknown action: %s", msg.Action)
	}
}
