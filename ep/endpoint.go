package ep

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/streadway/amqp"
)

type EpError struct {
	Name string
	Err  error
}

func NewEndpoint(ch *amqp.Channel, cfg EndpointConfig, epErrs chan EpError) *Endpoint {

	ep := &Endpoint{
		exit:     make(chan bool),
		exitResp: make(chan bool),
		errs:     epErrs,
	}
	ep.configure(ch, cfg)
	return ep
}

type Endpoint struct {
	Ch       *amqp.Channel
	Config   EndpointConfig
	exit     chan bool
	exitResp chan bool
	errs     chan EpError
}

func (e *Endpoint) Start() error {
	msgs, err := e.bindToRabbit()
	if err != nil {
		return err
	}

	go e.processMsgs(msgs, e.Config)

	return nil
}

func (e *Endpoint) Reload(ch *amqp.Channel, cfg EndpointConfig) error {
	log.Printf("Reloading endpoint %s", e.Config.Name)
	e.Stop()
	e.configure(ch, cfg)
	err := e.Start()
	log.Printf("Reloaded endpoint %s", e.Config.Name)
	return err
}

func (e *Endpoint) Stop() error {
	log.Printf("Stopping endpoint %s", e.Config.Name)

	e.exit <- true
	<-e.exitResp

	log.Printf("Stopped endpoint %s", e.Config.Name)
	return nil
}

func (e *Endpoint) configure(ch *amqp.Channel, cfg EndpointConfig) {
	e.Ch = ch
	e.Config = cfg
}

func (e *Endpoint) bindToRabbit() (<-chan amqp.Delivery, error) {
	log.Printf("Binding to Queue '%s'", e.Config.QueueName)

	q, err := e.Ch.QueueDeclare(
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

	err = e.Ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return nil, err
	}

	msgs, err := e.Ch.Consume(
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
func (e *Endpoint) processMsgs(msgs <-chan amqp.Delivery, cfg EndpointConfig) {
	defer e.Ch.Close()
	for {
		select {
		case <-e.exit:
			e.exitResp <- true
			return

		case d := <-msgs:
			log.Printf("Received a message on %s: %s", cfg.QueueName, string(d.Body))
			// nil msg start spewing when rabbit dies...
			//log.Printf("Wha? %+v", d)
			requeue, err := processMsg(d, cfg)
			if err != nil {
				log.Printf("%s: %s", cfg.Name, err)
				d.Nack(false, requeue)
			} else {
				log.Printf("%s: Message Processed", cfg.Name)
				d.Ack(false)
			}
			//e.errs <- EpError{Name: cfg.Name, Err: err}

		}
	}

}

func processMsg(d amqp.Delivery, cfg EndpointConfig) (bool, error) {
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
		// nack & don't requeue if endpoint responds with an error
		return false, fmt.Errorf("Code from '%s: %s' was '%d'", cfg.Method, url, r.StatusCode)
	}

	// ack
	return false, nil
}

func okStatus(code int) bool {
	return code == 200
}
