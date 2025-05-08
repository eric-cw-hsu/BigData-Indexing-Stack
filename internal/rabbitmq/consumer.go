package rabbitmq

import (
	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp091.Connection
	ch        *amqp091.Channel
	queueName string
}

func NewConsumer(rabbitMQUri, queueName string) (*Consumer, error) {
	conn, err := amqp091.Dial(rabbitMQUri)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &Consumer{
		conn:      conn,
		ch:        ch,
		queueName: queueName,
	}, nil
}

func (c *Consumer) Consume(handler func([]byte)) error {
	msgs, err := c.ch.Consume(
		c.queueName,
		"",
		true,  // autoAck
		false, // not exclusive
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for msg := range msgs {
		go handler(msg.Body)
	}
	return nil
}

func (c *Consumer) Close() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
