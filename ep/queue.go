package ep

import (
	"fmt"

	"github.com/benschw/chinchilla/config"
	"github.com/streadway/amqp"
)

const DefaultQueueType = "DefaultWorker"

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

func NewQueueRegistry() *QRegistry {
	return &QRegistry{
		DefaultKey: DefaultQueueType,
		reg:        make(map[string]Queue),
	}
}

type QRegistry struct {
	DefaultKey string
	reg        map[string]Queue
}

func (r *QRegistry) Add(key string, q Queue) *QRegistry {
	r.reg[key] = q
	return r
}

func (r *QRegistry) Get(key string) (Queue, error) {
	if key == "" {
		return r.Get(r.DefaultKey)
	}
	q, ok := r.reg[key]
	if !ok {
		return nil, fmt.Errorf("Queue strategy labeled '%s' doesn't exist", key)
	}
	return q, nil
}

// Set up in init
var queueReg *QRegistry

// add queue type to global registry
func RegisterQueueType(key string, q Queue) {
	queueReg.Add(key, q)
}

func QueueRegistry() *QRegistry {
	return queueReg
}

func init() {
	queueReg = NewQueueRegistry()
}
