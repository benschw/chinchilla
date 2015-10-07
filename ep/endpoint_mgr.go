package ep

import (
	"fmt"
	"log"
	"sync"

	"github.com/benschw/chinchilla/config"
	"github.com/streadway/amqp"
)

func NewEndpointMgr() *EndpointMgr {
	return &EndpointMgr{
		eps: make(map[string]*Endpoint),
	}
}

type EndpointMgr struct {
	eps map[string]*Endpoint
}

func (m *EndpointMgr) RestartAllEndpoints(conn *amqp.Connection) {
	log.Printf("Endpoints Reloading")

	var done sync.WaitGroup
	for _, ep := range m.eps {
		done.Add(1)
		go func(ep *Endpoint) {
			if err := m.RestartEndpoint(conn, ep.Config); err != nil {
				// @todo errors here might put things in a bad state
				log.Println(err)
			}
			done.Done()
		}(ep)
	}
	done.Wait()

	log.Printf("Endpoints Reloaded")
}

// Start endpoint, stopping first if it was already running. update index
func (m *EndpointMgr) RestartEndpoint(conn *amqp.Connection, cfg config.EndpointConfig) error {
	if old, ok := m.eps[cfg.Name]; ok {
		old.Stop()
		delete(m.eps, cfg.Name)
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	q, err := QueueRegistry().Get(cfg.QueueType)
	if err != nil {
		return err
	}
	ep, err := New(ch, cfg, q)
	if err != nil {
		return err
	}
	m.eps[cfg.Name] = ep
	return nil
}

func (m *EndpointMgr) StopAllEndpoints() {
	log.Printf("Endpoints Stopping")
	var done sync.WaitGroup
	for _, ep := range m.eps {
		done.Add(1)
		go func(ep *Endpoint) {
			if err := m.StopEndpoint(ep.Config.Name); err != nil {
				// @todo errors here might put things in a bad state
				log.Println(err)
			}
			done.Done()
		}(ep)
	}
	done.Wait()
	log.Printf("Endpoints Stopped")
	return
}

// Stop endpoint and remove from index
func (m *EndpointMgr) StopEndpoint(name string) error {
	old, ok := m.eps[name]
	if !ok {
		return fmt.Errorf("%s not present, can't stop", name)
	}
	old.Stop()
	delete(m.eps, name)
	return nil
}
