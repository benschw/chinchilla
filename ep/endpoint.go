package ep

import (
	"log"

	"github.com/streadway/amqp"
)

func NewEndpoint(ch *amqp.Channel, cfg EndpointConfig) (*Endpoint, error) {

	return &Endpoint{
		Ch:       ch,
		Config:   cfg,
		exit:     make(chan bool),
		exitResp: make(chan bool),
	}, nil

}

type Endpoint struct {
	Ch       *amqp.Channel
	Config   EndpointConfig
	exit     chan bool
	exitResp chan bool
}

func (e *Endpoint) Start() error {

	log.Printf("Binding to Queue '%s'", e.Config.QueueName)

	q, err := e.Ch.QueueDeclare(
		e.Config.QueueName, // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	err = e.Ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	msgs, err := e.Ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	go e.processMsgs(msgs)
	return nil
}
func (e *Endpoint) processMsgs(msgs <-chan amqp.Delivery) {
	for {
		select {
		case <-e.exit:
			e.exitResp <- true
			return

		case d := <-msgs:
			log.Printf("Received a message on %s: %s", e.Config.QueueName, d.Body)
			d.Ack(false)
			log.Printf("Done")
		}
	}

}

func (e *Endpoint) Stop() {
	log.Printf("Stopping consuming from queue %s", e.Config.QueueName)
	defer e.Ch.Close()

	e.exit <- true
	<-e.exitResp

	log.Printf("Stopped consuming from queue %s", e.Config.QueueName)
}
