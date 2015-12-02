package repeater

import "github.com/streadway/amqp"

type RepeaterLib struct {
	conProvider ConnectionProvider
}

func NewRepeaterLib(conProvider ConnectionProvider) *RepeaterLib {
	return &RepeaterLib{conProvider: conProvider}
}

func (r *RepeaterLib) Repeat(d amqp.Delivery, dc string, ex string) error {
	conn, err := r.conProvider.GetConnection(dc)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	return publishMessage(ch, ex, d)
}

func publishMessage(ch *amqp.Channel, ex string, d amqp.Delivery) error {

	err := ch.ExchangeDeclare(
		ex,      // name
		"topic", // type
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return err
	}

	err = ch.Publish(
		ex,           // exchange
		d.RoutingKey, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: d.ContentType,
			Body:        d.Body,
		},
	)
	return err
}
