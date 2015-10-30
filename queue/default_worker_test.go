package queue

import (
	"fmt"
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
			"queuename": "foo.bar",
			"prefetch":  5,
		},
	}

	publisher := &ex.Publisher{
		Conn:   conn,
		Config: &epCfg,
	}

	worker := &DefaultWorker{}

	ch, _ := conn.Channel()
	defer ch.Close()

	for i := 0; i < 10; i++ {
		publisher.Publish(fmt.Sprintf("test default worker: #%d", i), "text/plain")
	}

	// when
	msgs, _ := worker.Consume(ch, epCfg)

	// then
	cnt := countMessages(msgs)

	assert.Equal(t, 10, cnt, "wrong number of msgs")
}

func countMessages(msgs <-chan amqp.Delivery) int {

	var cnt = 0
	for {
		select {
		case d, _ := <-msgs:
			d.Ack(false)
			if d.Body == nil {
				return cnt
			}
			cnt++
		case <-time.After(5 * time.Millisecond):
			return cnt
		}
	}
	return 0
}
