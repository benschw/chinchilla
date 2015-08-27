package ep

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/benschw/chinchilla/config"
	"github.com/streadway/amqp"
)

type EpError struct {
	Name string
	Err  error
}

func New(ch *amqp.Channel, cfg config.EndpointConfig) (*Endpoint, error) {

	ep := &Endpoint{
		exit:     make(chan bool),
		exitResp: make(chan bool),
		ch:       ch,
		Config:   cfg,
	}
	return ep, ep.start()
}

type Endpoint struct {
	ch       *amqp.Channel
	Config   config.EndpointConfig
	exit     chan bool
	exitResp chan bool
}

func (e *Endpoint) start() error {
	log.Printf("%s: Starting Endpoint", e.Config.Name)
	msgs, err := e.bindToRabbit()
	if err != nil {
		return err
	}

	go e.processMsgs(msgs, e.Config)

	return nil
}

func (e *Endpoint) Stop() {
	log.Printf("%s: Endpoint Stopping", e.Config.Name)

	// if we detected a bad connection and already closed down the consumer, `e.exit` will be closed
	defer func() {
		recover()
	}()

	e.exit <- true
	close(e.exit)

	<-e.exitResp

	log.Printf("%s: Endpoint Stopped", e.Config.Name)
}

func (e *Endpoint) bindToRabbit() (<-chan amqp.Delivery, error) {

	q, err := e.ch.QueueDeclare(
		e.Config.QueueName, // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return nil, err
	}

	err = e.ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return nil, err
	}

	msgs, err := e.ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}
func (e *Endpoint) processMsgs(msgs <-chan amqp.Delivery, cfg config.EndpointConfig) {
	defer e.ch.Close()
	for {
		select {
		case <-e.exit:
			e.exitResp <- true
			return

		case d, ok := <-msgs:
			if !ok {
				log.Printf("%s: delivery chan closed", cfg.Name)
				//e.errs <- EpError{Name: cfg.Name, Err: fmt.Errorf("%s delivery chan was closed unexpectedly", cfg.Name)}
				close(e.exit)
				return
			}

			log.Printf("Received a message on %s: %s", cfg.QueueName, string(d.Body))
			requeue, err := processMsg(d, cfg)
			if err != nil {
				log.Printf("%s: %s", cfg.Name, err)
				d.Nack(false, requeue)
			} else {
				log.Printf("%s: Message Processed", cfg.Name)
				d.Ack(false)
			}
		}
	}
}

func processMsg(d amqp.Delivery, cfg config.EndpointConfig) (bool, error) {
	url := cfg.ServiceHost + cfg.Uri

	req, err := http.NewRequest(cfg.Method, url, bytes.NewBuffer(d.Body))
	if err != nil {
		// nack & requeue if we can't build a request
		return true, err
	}
	req.Header.Set("Content-Type", d.ContentType)

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		// nack & requeue if request errors out
		return true, err
	}

	if r == nil {
		// nack & requeue if response is nil
		return true, fmt.Errorf("Response from '%s: %s' was nil", cfg.Method, url)
	}

	if !okStatus(r.StatusCode) {
		// nack & requeue if response code is ! 2xx
		return true, fmt.Errorf("Code from '%s: %s' was '%d'", cfg.Method, url, r.StatusCode)
	}

	// ack
	return false, nil
}

func okStatus(code int) bool {
	return code >= 200 && code < 300
}
