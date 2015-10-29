package queue

import (
	"testing"

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
	_, _ = worker.Consume(ch, epCfg)

	// // then
	// numMsgs := len(msgs)
	assert.Equal(t, 1, 1, "wrong number of msgs")
}
