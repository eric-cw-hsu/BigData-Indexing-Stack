package rabbitmq

import (
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn  *amqp091.Connection
	ch    *amqp091.Channel
	queue amqp091.Queue
}

func NewPublisher(rabbitMQUrl string, queueName string) (*Publisher, error) {
	conn, err := amqp091.Dial(rabbitMQUrl)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := ch.QueueDeclare(
		queueName,
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		conn:  conn,
		ch:    ch,
		queue: queue,
	}, nil
}

func (p *Publisher) Publish(message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return p.ch.Publish(
		"",
		p.queue.Name,
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *Publisher) Close() {
	if p.ch != nil {
		p.ch.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}
