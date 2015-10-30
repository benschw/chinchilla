package queue

import (
	"fmt"
	"testing"
	"time"

	"github.com/benschw/chinchilla/config"
	"github.com/benschw/chinchilla/example/ex"
	"github.com/stretchr/testify/assert"
)

func TestTopicConsume(t *testing.T) {
	// given
	epCfg := config.EndpointConfig{
		QueueConfig: map[interface{}]interface{}{
			"prefetch":     5,
			"topicname":    "foo.update",
			"exchangename": "demo",
		},
	}

	publisher := &ex.Publisher{
		Conn:   conn,
		Config: &epCfg,
	}

	topic := &Topic{}

	ch, _ := conn.Channel()
	defer ch.Close()

	for i := 0; i < 10; i++ {
		publisher.PublishTopic(fmt.Sprintf("test topic: #%d", i), "text/plain")
	}

	// when
	msgs, err := topic.Consume(ch, epCfg)

	// then
	assert.Nil(t, err)
	time.Sleep(2000 * time.Millisecond)
	cnt := countMessages(msgs)

	assert.Equal(t, cnt, 10, "wrong number of msgs")
}
