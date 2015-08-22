package ep

import (
	"bytes"
	"log"
	"net/http"

	"github.com/streadway/amqp"
)

func NewEndpoint(ch *amqp.Channel, cfg EndpointConfig) *Endpoint {

	ep := &Endpoint{
		exit:     make(chan bool),
		exitResp: make(chan bool),
	}
	ep.configure(ch, cfg)
	return ep
}

type Endpoint struct {
	Ch       *amqp.Channel
	Config   EndpointConfig
	exit     chan bool
	exitResp chan bool
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
			url := cfg.ServiceHost + cfg.Uri

			req, err := http.NewRequest(cfg.Method, url, bytes.NewBuffer(d.Body))
			if err != nil {
				// @todo Handle me!
				log.Println(err)
			}
			req.Header.Set("Content-Type", d.ContentType)

			r, err := http.DefaultClient.Do(req)
			if err != nil {
				// @todo Handle me!
				log.Println(err)
			}
			if r == nil {
				// @todo Handle me!
				log.Println("response is nil")
			} else {
				if r.StatusCode == 200 {
					d.Ack(false)
					log.Println("Done: OK")
				} else {
					d.Reject(false)
					log.Println("Done: FAIL")
				}
			}
		}
	}

}
