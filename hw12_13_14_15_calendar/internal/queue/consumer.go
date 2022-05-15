package queue

import (
	"fmt"

	"github.com/streadway/amqp"
)

var _ Consumer = (*AMQPConsumer)(nil)

type AMQPConsumer struct {
	tag       string
	queue     string
	conn      *AMQPConnection
	ch        *amqp.Channel
	closeChan <-chan *amqp.Error
}

func (c *AMQPConsumer) Consume(h MessageHandler) error {
	deliveries, err := c.ch.Consume(
		c.queue,
		c.tag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("amqp consumer consume: %w", err)
	}

	for {
		select {
		case <-c.closeChan:
			return nil
		case d := <-deliveries:
			h(&Message{
				Key:     d.RoutingKey,
				Payload: d.Body,
			})
			_ = d.Ack(false)
		}
	}
}
