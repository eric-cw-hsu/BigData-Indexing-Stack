package messagequeue

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	channel  *amqp091.Channel
	exchange string
}

func NewPublisher(channel *amqp091.Channel, exchange string) (*Publisher, error) {

	if err := channel.ExchangeDeclare(
		exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, err
	}

	return &Publisher{
		channel:  channel,
		exchange: exchange,
	}, nil
}

func (p *Publisher) PublishMessage(ctx context.Context, routingKey string, msg Message) error {
	data, err := msg.Encode()
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(
		ctx,
		p.exchange,
		routingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        data,
		},
	)
}
