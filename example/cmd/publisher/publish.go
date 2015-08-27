package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/benschw/chinchilla/config"
	"github.com/benschw/chinchilla/example/ex"
	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func main() {
	queueName := flag.String("queue", "demo.foo", "supply a queue to publish to")
	contentType := flag.String("content-type", "text/plain", "set the message content type")
	body := flag.String("body", "Hello World", "Set the message's body")
	flag.Parse()

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	p := &ex.Publisher{
		Conn: conn,
		Config: &config.EndpointConfig{
			Name:      "TestEndpoint",
			QueueName: *queueName,
		},
	}
	err = p.Publish(*body, *contentType)
	if err != nil {
		panic(err)
	}
}
