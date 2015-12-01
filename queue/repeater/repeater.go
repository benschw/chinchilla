package repeater

import (
	"fmt"
	"log"

	"github.com/benschw/chinchilla/config"
	"github.com/streadway/amqp"
)

type Repeater struct {
	Lib *RepeaterLib
}

func (p *Repeater) Deliver(d amqp.Delivery, cfg config.EndpointConfig) {
	queueName, ok := cfg.QueueConfig["queuename"].(string)
	if !ok {
		queueName = "(unknown)"
	}
	log.Printf("Received a message on %s: %s", queueName, string(d.Body))
	requeue, err := processMsg(d, cfg, p.Lib)
	if err != nil {
		log.Printf("%s: %s", cfg.Name, err)
		d.Nack(false, requeue)
	} else {
		log.Printf("%s: Message Processed", cfg.Name)
		d.Ack(false)
	}
}

func processMsg(d amqp.Delivery, cfg config.EndpointConfig, lib *RepeaterLib) (bool, error) {
	connName, ok := cfg.QueueConfig["connection"].(string)
	if !ok {
		return true, fmt.Errorf("forwarding connection not defined in config")
	}
	exchange, ok := cfg.QueueConfig["exchangeout"].(string)
	if !ok {
		return true, fmt.Errorf("forwarding exchange not defined in config")
	}

	if err := lib.Repeat(d, connName, exchange); err != nil {
		// nack/requeue
		return true, err
	}

	// ack
	return false, nil
}
