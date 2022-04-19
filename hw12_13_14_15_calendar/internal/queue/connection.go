package queue

import (
	"fmt"
	"strconv"

	"github.com/streadway/amqp"
)

var _ Queue = (*AMQPConnection)(nil)

type AMQPConnection struct {
	URI             string
	conn            *amqp.Connection
	consumerCounter int
}

func (c *AMQPConnection) Connect() error {
	var err error
	c.conn, err = amqp.Dial(c.URI)
	if err != nil {
		return fmt.Errorf("connection dial: %w", err)
	}

	return nil
}

func (c *AMQPConnection) channel() (*amqp.Channel, error) {
	ch, err := c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("connection new channel: %w", err)
	}

	return ch, nil
}

func (c *AMQPConnection) declareExchange(ch *amqp.Channel, exchange string) error {
	if err := ch.ExchangeDeclare(
		exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("connection exchange `%s` declaration: %w", exchange, err)
	}

	return nil
}

func (c *AMQPConnection) declareQueue(ch *amqp.Channel, queue string) error {
	if _, err := ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("connection queue `%s` declaration: %w", queue, err)
	}

	return nil
}

func (c *AMQPConnection) CreateProducer(exchange string) (Producer, error) {
	ch, err := c.channel()
	if err != nil {
		return nil, fmt.Errorf("connection create pruducer: %w", err)
	}
	if err := c.declareExchange(ch, exchange); err != nil {
		return nil, fmt.Errorf("connection create pruducer: %w", err)
	}

	return &AMQPProducer{
		exchange: exchange,
		ch:       ch,
		conn:     c,
	}, nil
}

func (c *AMQPConnection) CreateConsumer(exchange, queue, key string) (Consumer, error) {
	ch, err := c.channel()
	if err != nil {
		return nil, fmt.Errorf("connection create consumer: %w", err)
	}

	if err := c.declareExchange(ch, exchange); err != nil {
		return nil, fmt.Errorf("connection create consumer: %w", err)
	}

	if err := c.declareQueue(ch, queue); err != nil {
		return nil, fmt.Errorf("connection create consumer: %w", err)
	}

	if err := ch.QueueBind(queue, key, exchange, true, nil); err != nil {
		return nil, fmt.Errorf("connection create consumer: queue binding: %w", err)
	}

	closeChan := make(chan *amqp.Error)
	c.conn.NotifyClose(closeChan)

	return &AMQPConsumer{
		c.createConsumerTag(), queue, c, ch, closeChan,
	}, nil
}

func (c *AMQPConnection) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("amqp connection close: %w", err)
		}
	}

	return nil
}

func (c *AMQPConnection) createConsumerTag() string {
	c.consumerCounter++

	return "consumer" + strconv.Itoa(c.consumerCounter)
}

func (c *AMQPConnection) isClosed() bool {
	return c.conn.IsClosed()
}
