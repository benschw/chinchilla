package queue

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/benschw/chinchilla/config"
	"github.com/benschw/chinchilla/example/ex"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
func publishABunchOfStuff(exch string, conn *amqp.Connection) {
	publisher := &ex.Publisher{
		Conn: conn,
		Config: &config.EndpointConfig{
			QueueConfig: map[interface{}]interface{}{
				"topicname":    "baz.update",
				"exchangename": exch,
			},
		},
	}
	publisher2 := &ex.Publisher{
		Conn: conn,
		Config: &config.EndpointConfig{
			QueueConfig: map[interface{}]interface{}{
				"topicname":    "baz.add",
				"exchangename": exch,
			},
		},
	}
	for i := 0; i < 10; i++ {
		publisher.PublishTopic(fmt.Sprintf("update msg: #%d", i), "text/plain")
		publisher2.PublishTopic(fmt.Sprintf("add msg: #%d", i), "text/plain")
	}
}

func TestTopicConsumeGlob(t *testing.T) {
	// given
	exch := RandomString(10)

	consumerCfg := config.EndpointConfig{
		QueueConfig: map[interface{}]interface{}{
			"queuename":    exch + "q",
			"topicname":    "baz.*",
			"exchangename": exch,
		},
	}

	topic := &Topic{}

	ch, _ := conn.Channel()
	defer ch.Close()

	// when
	msgs, err := topic.Consume(ch, consumerCfg)

	publishABunchOfStuff(exch, conn)

	// then
	assert.Nil(t, err)
	cnt := countMessages(msgs)

	assert.Equal(t, 20, cnt, "wrong number of msgs")
}

func TestTopicConsumeNegative(t *testing.T) {
	// given
	exch := RandomString(10)

	consumerCfg := config.EndpointConfig{
		QueueConfig: map[interface{}]interface{}{
			"queuename":    exch + "q",
			"topicname":    "iuweoiruwlekrjlk",
			"exchangename": exch,
		},
	}

	topic := &Topic{}

	ch, _ := conn.Channel()
	defer ch.Close()

	// when
	msgs, err := topic.Consume(ch, consumerCfg)

	publishABunchOfStuff(exch, conn)

	// then
	assert.Nil(t, err)
	cnt := countMessages(msgs)

	assert.Equal(t, 0, cnt, "wrong number of msgs")
}

func TestTopicConsumeFiltered(t *testing.T) {
	// given
	exch := RandomString(10)

	consumerCfg := config.EndpointConfig{
		QueueConfig: map[interface{}]interface{}{
			"queuename":    exch + "q",
			"topicname":    "baz.add",
			"exchangename": exch,
		},
	}

	topic := &Topic{}

	ch, _ := conn.Channel()
	defer ch.Close()

	// when
	msgs, err := topic.Consume(ch, consumerCfg)

	publishABunchOfStuff(exch, conn)

	// then
	assert.Nil(t, err)
	cnt := countMessages(msgs)

	assert.Equal(t, 10, cnt, "wrong number of msgs")
}
