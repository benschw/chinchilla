package queue

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/benschw/chinchilla/config"
	"github.com/streadway/amqp"
)

type DefaultDeliverer struct {
}

func (p *DefaultDeliverer) Deliver(d amqp.Delivery, cfg config.EndpointConfig) {
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

func processMsg(d amqp.Delivery, cfg config.EndpointConfig) (bool, error) {
	url, err := cfg.Url()
	if err != nil {
		// nack & requeue when there is a problem discovering url
		return true, err
	}

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
	defer r.Body.Close()

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