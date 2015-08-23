package main

import (
	"fmt"

	"testing"

	"github.com/benschw/chinchilla/ep"
	"github.com/benschw/chinchilla/example/ex"
	"github.com/benschw/opin-go/rando"
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
func GetPublisher(cfg *ep.EndpointConfig) *ex.Publisher {
	p := &ex.Publisher{
		Conn:   conn,
		Config: cfg,
	}
	return p
}

func GetServices() (*ep.Manager, *ex.Server, *ex.Publisher, *ex.Publisher) {
	port := uint16(rando.Port())

	server := ex.NewServer(fmt.Sprintf(":%d", port))

	epCfg := ep.EndpointConfig{
		Name:        "Foo",
		QueueName:   "test.foo",
		ServiceHost: fmt.Sprintf("http://localhost:%d", port),
		Uri:         "/foo",
		Method:      "POST",
	}
	epCfg2 := ep.EndpointConfig{
		Name:        "Bar",
		QueueName:   "test.bar",
		ServiceHost: fmt.Sprintf("http://localhost:%d", port),
		Uri:         "/bar",
		Method:      "POST",
	}

	p := GetPublisher(&epCfg)
	p2 := GetPublisher(&epCfg2)

	cfgUpdates := make(chan ep.ConfigUpdate)
	cfgUpdates <- ep.ConfigUpdate{T: ep.ConfigUpdateUpdate, Config: epCfg}
	cfgUpdates <- ep.ConfigUpdate{T: ep.ConfigUpdateUpdate, Config: epCfg2}

	epSvc := ep.NewManager(cfgUpdates)
	return epSvc, server, p, p2
}

func testPublish(t *testing.T) {
	// setup
	eps, server, p, _ := GetServices()
	go server.Start()

	go eps.Run()

	body := "Hello World"

	// when
	err := p.Publish(body, "text/plain")

	server.Stop()
	eps.Stop()

	// then
	assert.Nil(t, err)
	assert.Equal(t, 1, len(server.H.Stats["Foo"]), "wrong number of stats")
	assert.Equal(t, body, server.H.Stats["Foo"][0], "body not what expected")
}
func TestPublishLotsAndLots(t *testing.T) {
	// setup
	eps, server, p, p2 := GetServices()

	go server.Start()

	go eps.Run()

	body := "Hello World"

	// when
	for i := 0; i < 500; i++ {
		err := p.Publish(body, "text/plain")
		assert.Nil(t, err)

		err = p2.Publish(body, "text/plain")
		assert.Nil(t, err)

	}
	server.Stop()
	eps.Stop()

	// then
	assert.Equal(t, 500, len(server.H.Stats["Foo"]), "wrong number of stats")
	assert.Equal(t, 500, len(server.H.Stats["Bar"]), "wrong number of stats")
}
