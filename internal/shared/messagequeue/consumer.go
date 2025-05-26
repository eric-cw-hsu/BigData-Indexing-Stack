package messagequeue

import (
	"encoding/json"

	"eric-cw-hsu.github.io/internal/shared/logger"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type HandlerFunc func(body json.RawMessage) error

type Consumer struct {
	channel    *amqp091.Channel
	queueName  string
	handlerMap map[string]HandlerFunc
}

func NewConsumer(channel *amqp091.Channel, exchange, queueName string, routingKeys ...string) (*Consumer, error) {
	if err := channel.ExchangeDeclare(exchange, "direct", true, false, false, false, nil); err != nil {
		return nil, err
	}

	q, err := channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	for _, key := range routingKeys {
		if err := channel.QueueBind(q.Name, key, exchange, false, nil); err != nil {
			return nil, err
		}
	}

	return &Consumer{
		channel:    channel,
		queueName:  queueName,
		handlerMap: make(map[string]HandlerFunc),
	}, nil
}

func (c *Consumer) RegisterHandler(msgType string, handler HandlerFunc) {
	c.handlerMap[msgType] = handler
}

func (c *Consumer) Start() error {
	msgs, err := c.channel.Consume(
		c.queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			var baseMsg BaseMessage
			if err := json.Unmarshal(d.Body, &baseMsg); err != nil {
				logger.Logger.Error("Failed to unmarshal message", zap.Error(err))
				continue
			}

			handler, ok := c.handlerMap[baseMsg.Type]
			if !ok {
				logger.Logger.Error("No handler found for message type", zap.String("type", baseMsg.Type))
				continue
			}

			if err := handler(baseMsg.Body); err != nil {
				logger.Logger.Error("Failed to handle message", zap.Error(err))
				continue
			}
		}
	}()

	return nil
}

func (c *Consumer) Close() {
	logger.Logger.Info("Closing consumer channel")
	if err := c.channel.Close(); err != nil {
		logger.Logger.Error("Failed to close consumer channel", zap.Error(err))
	}

	logger.Logger.Info("Consumer channel closed")
}
