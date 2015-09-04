package queue

import (
	"fmt"

	"github.com/benschw/chinchilla/config"
	"github.com/streadway/amqp"
)

type DefaultWorker struct {
}

func (d *DefaultWorker) Consume(ch *amqp.Channel, cfg config.EndpointConfig) (<-chan amqp.Delivery, error) {
	queueName, ok := cfg.QueueConfig["queuename"].(string)
	if !ok {
		return nil, fmt.Errorf("unable to parse queuename from config")
	}

	prefetch, ok := cfg.QueueConfig["prefetch"].(int)
	if !ok || prefetch < 1 {
		prefetch = 1
	}

	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, err
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
