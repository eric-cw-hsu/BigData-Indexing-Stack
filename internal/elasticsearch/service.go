package elasticsearch

import (
	"log"

	"eric-cw-hsu.github.io/internal/elasticsearch/mappings"
	"eric-cw-hsu.github.io/internal/rabbitmq"
)

func Start(consumer *rabbitmq.Consumer, client *Client) {
	if err := client.InitIndex(mappings.GetPlanMapping()); err != nil {
		log.Fatalf("Failed to initialize index: %v", err)
	}

	go func() {
		consumer.Consume(
			func(body []byte) {
				ProcessRabbitMQPlanMessage(client, body)
			},
		)
	}()

	log.Println("Elasticsearch service started")
}
