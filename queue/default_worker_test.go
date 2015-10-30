package queue

import (
	"fmt"
	"testing"

	"github.com/benschw/chinchilla/config"
	"github.com/benschw/chinchilla/example/ex"
	"github.com/stretchr/testify/assert"
)

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
