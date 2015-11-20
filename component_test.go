package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"testing"

	"github.com/benschw/chinchilla/config"
	"github.com/benschw/chinchilla/ep"
	"github.com/benschw/chinchilla/example/ex"
	_ "github.com/benschw/chinchilla/queue"
	"github.com/benschw/opin-go/rando"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

var conn *amqp.Connection

type Msg struct {
	Message string
}

func init() {
	c, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	conn = c
}
func GetPublisher(cfg *config.EndpointConfig) *ex.Publisher {
	p := &ex.Publisher{
		Conn:   conn,
		Config: cfg,
	}
	return p
}

func GetServices() (*ep.EndpointApp, *ex.Server, *ex.Publisher, *ex.Publisher) {
	port := uint16(rando.Port())

	server := ex.NewServer(fmt.Sprintf(":%d", port))

	epCfg := config.EndpointConfig{
		Name:        "Foo",
		ServiceHost: fmt.Sprintf("http://localhost:%d", port),
		Uri:         "/foo",
		Method:      "POST",
		QueueConfig: map[interface{}]interface{}{
			"queuename": "test.foo",
		},
	}
	epCfg2 := config.EndpointConfig{
		Name:        "Bar",
		ServiceHost: fmt.Sprintf("http://localhost:%d", port),
		Uri:         "/bar",
		Method:      "POST",
		QueueConfig: map[interface{}]interface{}{
			"queuename": "test.bar",
		},
	}

	p := GetPublisher(&epCfg)
	p2 := GetPublisher(&epCfg2)

	repo := &config.StaticRepo{
		Address: config.RabbitAddress{
			User:     "guest",
			Password: "guest",
			Host:     "localhost",
			Port:     5672,
		},
		Endpoints: []config.EndpointConfig{epCfg, epCfg2},
	}

	mgr := ep.NewApp(repo, repo)
	return mgr, server, p, p2
}

func TestPublish(t *testing.T) {
	// setup
	mgr, server, p, _ := GetServices()
	go server.Start()
	go mgr.Run()

	// wait for queue creation to prevent race condition... do this better
	time.Sleep(200 * time.Millisecond)

	api := &Msg{Message: "Hello World"}
	apiB, _ := json.Marshal(api)
	apiStr := string(apiB)

	// when
	err := p.Publish(apiStr, "application/json")
	assert.Nil(t, err)

	time.Sleep(200 * time.Millisecond)

	mgr.Stop()
	server.Stop()

	// then

	statLen := len(server.H.Stats["Foo"])
	assert.Equal(t, 1, statLen, "wrong number of stats")
	if statLen > 0 {
		foundApi := &Msg{}
		err := json.Unmarshal([]byte(server.H.Stats["Foo"][0]), foundApi)
		assert.Nil(t, err, "err should be nil")

		assert.True(t, reflect.DeepEqual(api, foundApi), fmt.Sprintf("\n   %+v\n!= %+v", api, foundApi))
	}
}
func TestPublishLotsAndLots(t *testing.T) {
	// setup
	mgr, server, p, p2 := GetServices()
	go server.Start()
	go mgr.Run()

	body := "Hello World"

	// when
	for i := 0; i < 100; i++ {
		err := p.Publish(body, "text/plain")
		assert.Nil(t, err)

		err = p2.Publish(body, "text/plain")
		assert.Nil(t, err)

	}
	server.Stop()
	mgr.Stop()

	// then
	assert.Equal(t, 100, len(server.H.Stats["Foo"]), "wrong number of stats")
	assert.Equal(t, 100, len(server.H.Stats["Bar"]), "wrong number of stats")
}

func TestMessageHeaders(t *testing.T) {
	// setup
	mgr, server, p, _ := GetServices()
	go server.Start()
	go mgr.Run()

	body := "Hello World"

	// when
	err := p.Publish(body, "text/plain")
	assert.Nil(t, err)

	time.Sleep(3 * time.Second)

	server.Stop()
	mgr.Stop()

	// then
	timestamp, _ := time.Parse("2006-01-02 15:04:05", server.H.Request.Header.Get("X-Timestamp"))

	assert.Equal(t, "foo.poo", server.H.Request.Header.Get("X-reply_to"), "reply_to is wrong")
	assert.True(t, time.Now().Second()-timestamp.Second() <= 3, "wrong timestamp")
	assert.Equal(t, "test.foo", server.H.Request.Header.Get("X-routing_key"), "wrong routing key")
	assert.Equal(t, ex.MessageId, server.H.Request.Header.Get("X-message_id"), "wrong message id")
}
