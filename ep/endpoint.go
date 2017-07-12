package ep

import (
	"log"
	"sync"

	"github.com/benschw/chinchilla/config"
	"github.com/streadway/amqp"
)

func New(ch *amqp.Channel, cfg config.EndpointConfig, s *Strategy) (*Endpoint, error) {

	ep := &Endpoint{
		exit:     make(chan struct{}),
		exitResp: make(chan struct{}),
		ch:       ch,
		Config:   cfg,
		Strategy: s,
	}
	return ep, ep.start()
}

type Endpoint struct {
	ch       *amqp.Channel
	Config   config.EndpointConfig
	exit     chan struct{}
	exitResp chan struct{}
	Strategy *Strategy
}

func (e *Endpoint) start() error {
	log.Printf("%s: Starting Endpoint", e.Config.Name)
	msgs, err := e.Strategy.Consume(e.ch, e.Config)
	if err != nil {
		return err
	}

	go e.processMsgs(msgs, e.Config)

	return nil
}

func (e *Endpoint) Stop() {
	log.Printf("%s: Endpoint Stopping", e.Config.Name)

	defer func() {
		if x := recover(); x != nil {
			log.Printf("%s: Recovering after trying to stop a stopped endpoint", e.Config.Name)
		}
	}()

	close(e.exit)
	<-e.exitResp

	log.Printf("%s: Endpoint Stopped", e.Config.Name)
}

func (e *Endpoint) processMsgs(msgs <-chan amqp.Delivery, cfg config.EndpointConfig) {
	defer e.ch.Close()

	var wg sync.WaitGroup
	defer func() {
		wg.Wait()
		close(e.exitResp)
	}()
	for {
		select {
		case <-e.exit:
			return

		case d, ok := <-msgs:
			if !ok {
				log.Printf("%s: delivery chan closed", cfg.Name)
				close(e.exit)
				return
			}
			wg.Add(1)
			go func(d amqp.Delivery, cfg config.EndpointConfig) {
				defer wg.Done()
				e.Strategy.Deliver(d, cfg)
			}(d, cfg)
		}
	}
}
