package ep

import (
	"github.com/benschw/chinchilla/config"
	"github.com/streadway/amqp"
)

const (
	DefaultConsumerStrategy = "DefaultConsumer"
	DefaultDeliveryStrategy = "DefaultDeliverer"
)

// Container for configured Consume and Deliver implementations
type Strategy struct {
	C MsgConsumer
	D MsgDeliverer
}

// Facade for MsgConsumer.Consume
func (q *Strategy) Consume(ch *amqp.Channel, cfg config.EndpointConfig) (<-chan amqp.Delivery, error) {
	return q.C.Consume(ch, cfg)
}

// Facade for MsgDeliverer.Deliver
func (q *Strategy) Deliver(d amqp.Delivery, cfg config.EndpointConfig) {
	q.D.Deliver(d, cfg)
}

// Interface to support pluggable queue configurations
type MsgConsumer interface {
	Consume(*amqp.Channel, config.EndpointConfig) (<-chan amqp.Delivery, error)
}

// Interface to support pluggable delivery strategies
type MsgDeliverer interface {
	Deliver(d amqp.Delivery, cfg config.EndpointConfig)
}
