package queue

import (
	"fmt"

	"github.com/streadway/amqp"
)

var _ Producer = (*AMQPProducer)(nil)

type AMQPProducer struct {
	exchange string
	conn     *AMQPConnection
	ch       *amqp.Channel
}

func (p *AMQPProducer) Publish(m *Message) error {
	if p.conn.isClosed() {
		return ErrConnectionClosed
	}

	if err := p.ch.Publish(
		p.exchange,
		m.Key,
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "application/json",
			ContentEncoding: "",
			Body:            m.Payload,
			DeliveryMode:    amqp.Transient,
		}); err != nil {
		return fmt.Errorf("amqp producer publish: %w", err)
	}

	return nil
}
