package queue

import (
	"errors"
)

var ErrConnectionClosed = errors.New("connection is closed")

type Message struct {
	Key     string
	Payload []byte
}

type MessageHandler func(m *Message)

type Queue interface {
	Connect() error
	Close() error
	CreateProducer(exchange string) (Producer, error)
	CreateConsumer(exchange, queue, key string) (Consumer, error)
}

type Producer interface {
	Publish(m *Message) error
}

type Consumer interface {
	Consume(h MessageHandler) error
}

func New(uri string) Queue {
	return &AMQPConnection{
		URI: uri,
	}
}
