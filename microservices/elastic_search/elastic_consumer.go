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

				if action, ok := event["action"].(string); ok {
					switch action {
					case "create", "update":

						if data, ok := event["data"].(map[string]interface{}); ok {
							dataBytes, _ := json.Marshal(data)
							ec.elasticService.IndexToElastic("plan_index", event["key"].(string), dataBytes)
						} else {
							ec.logger.Error("Event missing data field")
						}
					case "delete":
						if key, ok := event["key"].(string); ok {
							ec.elasticService.DeleteFromElastic("plan_index", key)
						} else {
							ec.logger.Error("Delete event missing key field")
						}
					default:
						ec.logger.Warn("Unknown event action: ", action)
					}
				} else {
					ec.logger.Error("Event missing action field")
				}
			case <-ctx.Done():
				ec.logger.Info("ElasticConsumer context cancelled")
				return
			}
		}
	}()

	return nil
}
