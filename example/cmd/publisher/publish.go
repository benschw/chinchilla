package main

import (
	"flag"
	"fmt"
	"log"
	"sync"

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
	runs := flag.Int("runs", 1, "msgs to publish")
	queueName := flag.String("queue", "", "supply a queue to publish to")
	topicName := flag.String("topic", "foo.update", "supply a topic to publish to")
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
			Name: "TestEndpoint",
			QueueConfig: map[interface{}]interface{}{
				"queuename":    *queueName,
				"topicname":    *topicName,
				"exchangename": "demo",
			},
		},
	}
	var done sync.WaitGroup
	for i := 0; i < *runs; i++ {
		done.Add(1)
		go func(i int) {
			if len(*queueName) > 0 {
				err = p.Publish(fmt.Sprintf("%s-%d", *body, i), *contentType)
			} else if len(*topicName) > 0 {
				err = p.PublishTopic(fmt.Sprintf("%s-%d", *body, i), *contentType)
			}

			if err != nil {
				panic(err)
			}
			done.Done()
		}(i)
	}
	done.Wait()
}
