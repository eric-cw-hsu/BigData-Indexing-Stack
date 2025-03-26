package queue

import (
	"fmt"

	"eric-cw-hsu.github.io/errors"
	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type RabbitMQQueue struct {
	connection *amqp091.Connection
	channel    *amqp091.Channel
	queueName  string
	logger     *logrus.Logger
}

// NewRabbitMQQueue creates a new RabbitMQ connection and declares a queue.
// It returns an error instead of calling fatal, so that the caller can decide how to handle failures.
func NewRabbitMQQueue(amqpURL, queueName string, logger *logrus.Logger) (*RabbitMQQueue, error) {
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ: ", err)
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		logger.Error("Failed to open a channel: ", err)
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		logger.Error("Failed to declare a queue: ", err)
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	return &RabbitMQQueue{
		connection: conn,
		channel:    ch,
		queueName:  queueName,
		logger:     logger,
	}, nil
}

// Publish sends a message to the declared queue.
func (rq *RabbitMQQueue) Publish(message []byte) error {
	err := rq.channel.Publish(
		"",
		rq.queueName,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        message,
		},
	)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrFailedToPublishMessage, err)
	}
	rq.logger.Infof("Message published to queue %s", rq.queueName)
	return nil
}

// Close cleanly closes the channel and connection.
func (rq *RabbitMQQueue) Close() error {
	if err := rq.channel.Close(); err != nil {
		return err
	}
	return rq.connection.Close()
}
