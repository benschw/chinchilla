package queue

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/benschw/chinchilla/config"
	"github.com/streadway/amqp"
)

type DefaultDeliverer struct {
}

func (p *DefaultDeliverer) Deliver(d amqp.Delivery, cfg config.EndpointConfig) {
	queueName, ok := cfg.QueueConfig["queuename"].(string)
	if !ok {
		queueName = "(unknown)"
	}
	log.Printf("Received a message on %s", queueName)
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
	timeoutSec, ok := cfg.QueueConfig["timeout"].(int)
	if !ok {
		timeoutSec = 60
	}

	url, err := cfg.Url()
	if err != nil {
		// nack & requeue when there is a problem discovering url
		return true, err
	}
	log.Println("url: " + url)
	req, err := http.NewRequest(cfg.Method, url, bytes.NewBuffer(d.Body))
	if err != nil {
		// nack & requeue if we can't build a request
		return true, err
	}
	req.Header.Set("Content-Type", d.ContentType)
	req.Header.Set("X-reply-to", d.ReplyTo)
	req.Header.Set("X-expiration", d.Expiration)
	req.Header.Set("X-message-id", d.MessageId)
	req.Header.Set("X-timestamp", d.Timestamp.Format("2006-01-02 15:04:05"))
	req.Header.Set("X-exchange", d.Exchange)
	req.Header.Set("X-routing-key", d.RoutingKey)

	timeout := time.Duration(time.Duration(timeoutSec) * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	r, err := client.Do(req)
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
