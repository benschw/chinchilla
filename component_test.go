package main

import (
	"fmt"
	"time"

	"testing"

	"github.com/benschw/chinchilla/config"
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
func GetPublisher(cfg *config.EndpointConfig) *ex.Publisher {
	p := &ex.Publisher{
		Conn:   conn,
		Config: cfg,
	}
	return p
}

func GetServices() (*ep.EndpointManager, *ex.Server, *ex.Publisher, *ex.Publisher) {
	port := uint16(rando.Port())

	server := ex.NewServer(fmt.Sprintf(":%d", port))

	epCfg := config.EndpointConfig{
		Name:        "Foo",
		QueueName:   "test.foo",
		ServiceHost: fmt.Sprintf("http://localhost:%d", port),
		Uri:         "/foo",
		Method:      "POST",
	}
	epCfg2 := config.EndpointConfig{
		Name:        "Bar",
		QueueName:   "test.bar",
		ServiceHost: fmt.Sprintf("http://localhost:%d", port),
		Uri:         "/bar",
		Method:      "POST",
	}

	p := GetPublisher(&epCfg)
	p2 := GetPublisher(&epCfg2)

	cfgMgr := config.NewWatcher(&config.StaticRepo{Endpoints: []config.EndpointConfig{epCfg, epCfg2}})

	ap := &config.StaticRepo{
		Address: config.RabbitAddress{
			User:     "guest",
			Password: "guest",
			Host:     "localhost",
			Port:     5672,
		},
	}
	mgr := ep.NewManager(ap, cfgMgr)
	return mgr, server, p, p2
}

func TestPublish(t *testing.T) {
	// setup
	mgr, server, p, _ := GetServices()
	go server.Start()
	go mgr.Run()

	body := "Hello World"

	// when
	err := p.Publish(body, "text/plain")
	assert.Nil(t, err)

	time.Sleep(200 * time.Millisecond)

	mgr.Stop()
	server.Stop()

	// then
	statLen := len(server.H.Stats["Foo"])
	assert.Equal(t, 1, statLen, "wrong number of stats")
	if statLen > 0 {
		assert.Equal(t, body, server.H.Stats["Foo"][0], "body not what expected")
	}
}
func TestPublishLotsAndLots(t *testing.T) {
	// setup
	mgr, server, p, p2 := GetServices()
	go server.Start()
	go mgr.Run()

	body := "Hello World"

	// when
	for i := 0; i < 500; i++ {
		err := p.Publish(body, "text/plain")
		assert.Nil(t, err)

		err = p2.Publish(body, "text/plain")
		assert.Nil(t, err)

	}
	server.Stop()
	mgr.Stop()

	// then
	assert.Equal(t, 500, len(server.H.Stats["Foo"]), "wrong number of stats")
	assert.Equal(t, 500, len(server.H.Stats["Bar"]), "wrong number of stats")
}
