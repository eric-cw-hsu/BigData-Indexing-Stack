package rabbitmq

import "github.com/rabbitmq/amqp091-go"

type MQConnection struct {
	Conn    *amqp091.Connection
	Channel *amqp091.Channel
}

func NewMQConnection(uri string) (*MQConnection, error) {
	conn, err := amqp091.Dial(uri)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &MQConnection{
		Conn:    conn,
		Channel: ch,
	}, nil
}

func (mq *MQConnection) Close() {
	if mq.Channel != nil {
		mq.Channel.Close()
	}

	if mq.Conn != nil {
		mq.Conn.Close()
	}
}
