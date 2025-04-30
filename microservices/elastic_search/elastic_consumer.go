package main

import (
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type ElasticConsumer struct {
	channel        *amqp091.Channel
	queueName      string
	elasticService *ElasticSearchService
	logger         *logrus.Logger
}

func NewElasticConsumer(ch *amqp091.Channel, queueName string, elasticService *ElasticSearchService, logger *logrus.Logger) *ElasticConsumer {
	ec := &ElasticConsumer{
		channel:        ch,
		queueName:      queueName,
		elasticService: elasticService,
		logger:         logger,
	}

	ec.StartConsuming(context.Background())

	return ec
}

func (ec *ElasticConsumer) StartConsuming(ctx context.Context) error {
	msgs, err := ec.channel.Consume(
		ec.queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ec.logger.Error("Failed to register consumer: ", err)
		return err
	}

	go func() {
		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					ec.logger.Warn("Message channel closed")
					return
				}
				ec.logger.Infof("Received a message: %s", msg.Body)
				var event map[string]interface{}
				if err := json.Unmarshal(msg.Body, &event); err != nil {
					ec.logger.Error("Failed to unmarshal event: ", err)
					continue
				}

				action, ok := event["action"].(string)
				if !ok {
					ec.logger.Error("Event missing action field")
					continue
				}

				data, ok := event["data"].(map[string]interface{})
				if !ok {
					ec.logger.Error("Event missing data field")
					continue
				}

				for _, value := range ec.splitParentChildren("plan", "", data) {
					parent := ""
					if value["join_field"].(map[string]interface{})["parent"] != nil {
						parent = value["join_field"].(map[string]interface{})["parent"].(string)
					}

					dataBytes, _ := json.Marshal(value)

					if action == "delete" {
						ec.elasticService.DeleteFromElastic(event["index"].(string), value["objectId"].(string))
					} else {
						ec.elasticService.IndexToElastic(event["index"].(string), value["objectId"].(string), dataBytes, parent)
					}
				}

			case <-ctx.Done():
				ec.logger.Info("ElasticConsumer context cancelled")
				return
			}
		}
	}()

	return nil
}

func (ec *ElasticConsumer) splitParentChildren(name, parent string, data map[string]interface{}) []map[string]interface{} {
	var result []map[string]interface{}
	id := data["objectId"].(string)

	joinField := map[string]interface{}{
		"name": name,
	}
	if parent != "" {
		joinField["parent"] = parent
	}

	current := map[string]interface{}{
		"join_field": joinField,
	}

	for key, value := range data {
		if value, ok := value.(map[string]interface{}); ok {
			result = append(
				result,
				ec.splitParentChildren(key, id, value)...,
			)
			continue
		}

		if value, ok := value.([]interface{}); ok {
			for _, v := range value {
				if v, ok := v.(map[string]interface{}); ok {
					result = append(
						result,
						ec.splitParentChildren(key, id, v)...,
					)
				}
			}
			continue
		}

		current[key] = value
	}

	result = append(result, current)

	return result
}
