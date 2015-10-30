package ex

import (
	"fmt"
	"github.com/benschw/chinchilla/config"
	"github.com/streadway/amqp"
	"log"
)

type Publisher struct {
	Conn   *amqp.Connection
	Config *config.EndpointConfig
}

func (p *Publisher) Publish(body string, contentType string) error {
	queueName, ok := p.Config.QueueConfig["queuename"].(string)
	if !ok {
		return fmt.Errorf("unable to parse queuename from config")
	}

	ch, err := p.Conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: contentType,
			Body:        []byte(body),
		})
	if err != nil {
		return err
	}
	log.Printf(" [x] Sent %s", body)
	return nil
}

func (p *Publisher) PublishTopic(body string, contentType string) error {
	topicName, ok := p.Config.QueueConfig["topicname"].(string)
	if !ok {
		return fmt.Errorf("unable to parse queuename from config")
	}

	exchangeName, ok := p.Config.QueueConfig["exchangename"].(string)
	if !ok {
		return fmt.Errorf("unable to parse exchangename from config")
	}

	ch, err := p.Conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)

	err = ch.Publish(
		exchangeName, // exchange
		topicName,    // routing key @todo what should this be?
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: contentType,
			Body:        []byte(body),
		})
	if err != nil {
		return err
	}
	log.Printf(" [x] Sent %s", body)
	return nil
}
