package ex

import (
	"github.com/benschw/chinchilla/config"
	"github.com/streadway/amqp"
)

type Publisher struct {
	Conn   *amqp.Connection
	Config *config.EndpointConfig
}

func (p *Publisher) Publish(body string, contentType string) error {
	ch, err := p.Conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.Publish(
		"",                 // exchange
		p.Config.QueueName, // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType: contentType,
			Body:        []byte(body),
		})
	if err != nil {
		return err
	}
	//	log.Printf(" [x] Sent %s", body)
	return nil
}
