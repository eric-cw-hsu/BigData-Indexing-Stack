package elasticsearch

import (
	"log"

	"eric-cw-hsu.github.io/internal/elasticsearch/mappings"
	"eric-cw-hsu.github.io/internal/shared/messagequeue"
)

func Start(consumer *messagequeue.Consumer, client *Client) {
	if err := client.InitIndex(mappings.GetPlanMapping()); err != nil {
		log.Fatalf("Failed to initialize index: %v", err)
	}

	consumer.RegisterHandler("plan.node.create", ProcessCreatePlanNode(client))
	consumer.RegisterHandler("plan.node.update", ProcessCreatePlanNode(client))
	consumer.RegisterHandler("plan.node.delete", ProcessDeletePlanNode(client))

	if err := consumer.Start(); err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}

	log.Println("Elasticsearch service started")
}
