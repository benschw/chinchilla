package queue

import (
	"log"
	"testing"
	"time"

	"github.com/benschw/chinchilla/config"
	"github.com/benschw/chinchilla/example/ex"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

var conn *amqp.Connection

func init() {
	c, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	conn = c
}

func TestDefaultWorkerConsume(t *testing.T) {
	// given
	epCfg := config.EndpointConfig{
		QueueConfig: map[interface{}]interface{}{
			"queuename": "foo.yahoo10",
			"prefetch":  5,
		},
	}

	publisher := &ex.Publisher{
		Conn:   conn,
		Config: &epCfg,
	}

	worker := &DefaultWorker{}

	ch, _ := conn.Channel()

	// when
	publisher.Publish("test default worker", "text/plain")
	publisher.Publish("test default worker 2", "text/plain")
	publisher.Publish("test default worker 3", "text/plain")
	time.Sleep(1000 * time.Millisecond)
	msgs, _ := worker.Consume(ch, epCfg)

	defer ch.Close()

	var cnt = 0

Loop:
	for {
		select {
		case d, _ := <-msgs:
			d.Ack(false)
			cnt++
			log.Println(cnt)
		default:
			log.Println("break")
			break Loop
		}
	}

	assert.Equal(t, cnt, 3, "wrong number of msgs")
}
