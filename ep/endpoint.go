package ep

import (
	"bytes"
	"log"
	"time"

	"github.com/streadway/amqp"
)

func NewEndpoint(conn *amqp.Connection, cfg EndpointConfig) *Endpoint {
	return &Endpoint{Conn: conn, Config: cfg}
}

type Endpoint struct {
	Conn   *amqp.Connection
	Config EndpointConfig
}

func (e *Endpoint) Run() error {
	log.Printf("Binding to Queue '%s'", e.Config.QueueName)
	ch, err := e.Conn.Channel()
	if err != nil {
		log.Println(err)
		panic(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
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

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Println(err)
		panic(err)
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
		log.Println(err)
		panic(err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message on %s: %s", e.Config.QueueName, d.Body)
			d.Ack(false)
			dot_count := bytes.Count(d.Body, []byte("."))
			t := time.Duration(dot_count)
			time.Sleep(t * time.Second)
			log.Printf("Done")
		}
	}()

	<-forever

	return nil
}
