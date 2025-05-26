package elasticsearch

import (
	"encoding/json"
	"log"

	"eric-cw-hsu.github.io/internal/rabbitmq/types"
	"eric-cw-hsu.github.io/internal/shared/messagequeue"
)

func ParseRabbitMQPlanMessage(body json.RawMessage) (types.PlanMessage, error) {
	var msg types.PlanMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return msg, err
	}
	return msg, nil
}

func ProcessCreatePlanNode(client *Client) messagequeue.HandlerFunc {
	return func(body json.RawMessage) error {
		planMessage, err := ParseRabbitMQPlanMessage(body)
		if err != nil {
			log.Printf("Failed to parse message: %v", err)
		}

		parentId, ok := planMessage.Data["parentId"].(string)
		if !ok {
			parentId = ""
		}

		client.IndexDocument(planMessage.Key, planMessage.Data, parentId)
		return nil
	}
}

func ProcessDeletePlanNode(client *Client) messagequeue.HandlerFunc {
	return func(body json.RawMessage) error {
		planMessage, err := ParseRabbitMQPlanMessage(body)
		if err != nil {
			log.Printf("Failed to parse message: %v", err)
		}

		client.DeleteDocument(planMessage.Key)
		return nil
	}
}
