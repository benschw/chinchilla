package ep

import (
	"fmt"

	"github.com/benschw/chinchilla/config"
	"github.com/streadway/amqp"
)

type Queue interface {
	MsgConsumer
	MsgDeliverer
}

// Interface to support pluggable queue configurations
type MsgConsumer interface {
	Consume(*amqp.Channel, config.EndpointConfig) (<-chan amqp.Delivery, error)
}

// Interface to support pluggable delivery strategies
type MsgDeliverer interface {
	Deliver(d amqp.Delivery, cfg config.EndpointConfig)
}

func NewQueueRegistry() *QueueRegistry {
	return &QueueRegistry{
		DefaultKey: "DefaultWorker",
		reg:        make(map[string]Queue),
	}
}

type QueueRegistry struct {
	DefaultKey string
	reg        map[string]Queue
}

func (r *QueueRegistry) Add(key string, q Queue) *QueueRegistry {
	r.reg[key] = q
	return r
}

func (r *QueueRegistry) Get(key string) (Queue, error) {
	if key == "" {
		return r.Get(r.DefaultKey)
	}
	q, ok := r.reg[key]
	if !ok {
		return nil, fmt.Errorf("Queue strategy labeled '%s' doesn't exist", key)
	}
	return q, nil
}
