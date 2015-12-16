package queue

import (
	"fmt"

	"github.com/benschw/chinchilla/config"
	"github.com/streadway/amqp"
)

type Topic struct {
}

func (d *Topic) Consume(ch *amqp.Channel, cfg config.EndpointConfig) (<-chan amqp.Delivery, error) {
	queueName, ok := cfg.QueueConfig["queuename"].(string)
	if !ok {
		return nil, fmt.Errorf("unable to parse queuename from config")
	}

	topicName, ok := cfg.QueueConfig["topicname"].(string)
	if !ok {
		return nil, fmt.Errorf("unable to parse topicname from config")
	}

	exchangeName, ok := cfg.QueueConfig["exchangename"].(string)
	if !ok {
		return nil, fmt.Errorf("unable to parse exchangename from config")
	}

	prefetch, ok := cfg.QueueConfig["prefetch"].(int)
	if !ok || prefetch < 1 {
		prefetch = 1
	}

	err := ch.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when usused
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

	err = ch.QueueBind(
		q.Name,       // queue name
		topicName,    // routing key
		exchangeName, // exchange
		false,
		nil,
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
