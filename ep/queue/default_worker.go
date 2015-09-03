package queue

import (
	"github.com/benschw/chinchilla/config"
	"github.com/streadway/amqp"
)

type DefaultWorker struct {
}

func (d *DefaultWorker) Consume(ch *amqp.Channel, cfg config.EndpointConfig) (<-chan amqp.Delivery, error) {
	q, err := ch.QueueDeclare(
		cfg.QueueName, // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return nil, err
	}

	prefetch := cfg.Prefetch
	if prefetch < 1 {
		prefetch = 1
	}
	err = ch.Qos(
		prefetch, // prefetch count
		0,        // prefetch size
		false,    // global
	)
	if err != nil {
		return nil, err
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}
