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

func GetPublisher(cfg *ep.EndpointConfig) *ex.Publisher {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	p := &ex.Publisher{
		Conn:   conn,
		Config: cfg,
	}
	return p
}

func GetServices() (*ep.Service, *ex.Server, *ex.Publisher) {
	port := uint16(rando.Port())

	server := ex.NewServer(fmt.Sprintf(":%d", port))

	epCfg := ep.EndpointConfig{
		Name:        "Foo",
		QueueName:   "test.foo",
		ServiceHost: fmt.Sprintf("http://localhost:%d", port),
		Uri:         "/foo",
		Method:      "POST",
	}
	cfg := ep.Config{Endpoints: []ep.EndpointConfig{
		epCfg,
	}}

	p := GetPublisher(&epCfg)

	epSvc := ep.New(cfg)
	return epSvc, server, p
}

func TestPublish(t *testing.T) {
	// Setup
	eps, server, p := GetServices()
	go server.Start()
	defer server.Stop()

	eps.Start()
	defer eps.Stop()

	// When

	err := p.Publish("Hello", "text/plain")

	// Then
	assert.Nil(t, err)
}
