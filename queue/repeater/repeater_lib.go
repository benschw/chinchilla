package repeater

import (
	"log"
	"time"

	"github.com/benschw/chinchilla/config"
	"github.com/benschw/chinchilla/ep"
	"github.com/streadway/amqp"
)

type RepeaterLib struct {
	add map[string]config.RabbitAddress
	chs map[string]chan PublishRequest
}

func NewRepeaterLib(arr []config.RabbitAddress) *RepeaterLib {
	l := &RepeaterLib{
		add: make(map[string]config.RabbitAddress),
		chs: make(map[string]chan PublishRequest),
	}

	for _, a := range arr {
		ch := make(chan PublishRequest)
		l.add[a.Name] = a
		l.chs[a.Name] = ch

		go repeaterDeMux(&ConnectionAddressProvider{add: a}, ch)
	}
	return l
}

func (r *RepeaterLib) Repeat(d amqp.Delivery, dc string, ex string) error {
	resp := make(chan error)
	req := PublishRequest{
		ex:   ex,
		d:    d,
		resp: resp,
	}
	r.chs[dc] <- req

	return <-resp
}

type PublishRequest struct {
	ex   string
	d    amqp.Delivery
	resp chan error
}

func repeaterDeMux(add config.RabbitAddressProvider, req chan PublishRequest) {
	var cErr chan *amqp.Error
	conn, connErr, err := ep.DialRabbit(add)
	if err != nil {
		log.Printf("Fatal Repeater Connection Error: %s", err)
		return
	}
	cErr = connErr
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Fatal Repeater Connection Error: %s", err)
		return
	}
	defer ch.Close()

	for {
		select {
		case err, ok := <-cErr:
			if err != nil {
				log.Printf("Connection Lost: %s", err)
			}
			if !ok {
				log.Printf("Waiting %d seconds before reconnect attempt", 5)
				time.Sleep(5 * time.Second)
			}
			conn, connErr, e := ep.DialRabbit(add)
			if e != nil {
				log.Printf("Can't Reconnect: %s", e)
				break
			}
			cErr = connErr
			ch, e = conn.Channel()
			if err != nil {
				log.Printf("Can't create channel: %s", e)
				log.Printf("Fatal Repeater Connection Error: %s", err)
				return
			}
			defer ch.Close()

		case r := <-req:
			r.resp <- publishMessage(ch, r.ex, r.d)
		}
	}
}

func publishMessage(ch *amqp.Channel, ex string, d amqp.Delivery) error {
	log.Printf("%s: %s", ex, string(d.Body[:]))

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
