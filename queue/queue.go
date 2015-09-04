package queue

import (
	"github.com/benschw/chinchilla/config"
	"github.com/benschw/chinchilla/ep"
	"github.com/streadway/amqp"
)

type Queue struct {
	C ep.MsgConsumer
	D ep.MsgDeliverer
}

func (q *Queue) Consume(ch *amqp.Channel, cfg config.EndpointConfig) (<-chan amqp.Delivery, error) {
	return q.C.Consume(ch, cfg)
}
func (q *Queue) Deliver(d amqp.Delivery, cfg config.EndpointConfig) {
	q.D.Deliver(d, cfg)
}
