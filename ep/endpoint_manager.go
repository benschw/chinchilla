package ep

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/benschw/chinchilla/config"
	"github.com/streadway/amqp"
)

type Trigger int

const (
	TriggerStop   = iota
	TriggerReload = iota
)

func NewManager(ap config.RabbitAddressProvider, epp config.EndpointsProvider) *EndpointManager {
	return &EndpointManager{
		ap:        ap,
		epp:       epp,
		eps:       make(map[string]*Endpoint),
		ttl:       5,
		connRetry: 2,
		ex:        make(chan struct{}, 1),
	}
}

type EndpointManager struct {
	ap        config.RabbitAddressProvider
	epp       config.EndpointsProvider
	conn      *amqp.Connection
	connErr   chan *amqp.Error
	eps       map[string]*Endpoint
	ttl       int
	connRetry int
	ex        chan struct{}
}

func (m *EndpointManager) connect() error {
	conn, connErr, err := DialRabbit(m.ap)
	if err != nil {
		return err
	}
	m.conn = conn
	m.connErr = connErr
	return nil
}

func (m *EndpointManager) Run() error {
	log.Println("Starting...")

	if err := m.connect(); err != nil {
		return err
	}
	defer m.conn.Close()

	cfgWatcher := config.NewWatcher(m.epp, m.ttl)
	defer cfgWatcher.Stop()

	sigWatcher := NewSignalWatcher()
	defer sigWatcher.Stop()

	// main control flow
	for {
		select {
		// `Stop` func signals this chan to break out of the main loop
		case <-m.ex:
			err := m.stopAllEndpoints()
			log.Println("Leaving control loop")
			return err

		// If a signal is caught, either shutdown or reload gracefully
		case t := <-sigWatcher.T:
			switch t {
			case TriggerStop:
				m.Stop()
			case TriggerReload:
				m.Reload()
			}

		// If connection is lost, keep trying to reconnect forever
		case err, ok := <-m.connErr:
			if err != nil {
				log.Printf("Connection Lost: %s", err)
			}
			if !ok {
				log.Printf("Waiting %d seconds before reconnect attempt", m.connRetry)
				time.Sleep(time.Duration(m.connRetry) * time.Second)
			}
			if err := m.connect(); err != nil {
				log.Printf("Can't Reconnect: %s", err)
				continue
			}
			m.reloadEndpoints()

		// Handle incoming config updates
		case cfgU := <-cfgWatcher.Updates:
			switch cfgU.T {
			case config.EndpointUpdate:
				if err := m.restartEndpoint(cfgU.Config); err != nil {
					log.Printf("%s: problem starting/reloading: %s", cfgU.Config.Name, err)
				}
			case config.EndpointDelete:
				if err := m.stopEndpoint(cfgU.Name); err != nil {
					log.Printf("%s: problem stopping: %s", cfgU.Name, err)
				}
			}
		}
	}
	return nil
}

func (m *EndpointManager) Reload() {
	log.Printf("Endpoint Manager Reloading")
	if err := m.connect(); err != nil {
		// just log out the problem and don't try to recover
		log.Printf("Couldn't reconnect: %s", err)
		return
	}
	m.reloadEndpoints()
	log.Printf("Endpoint Manager Reloaded")
}
func (m *EndpointManager) Stop() {
	log.Printf("Stopping %d Endpoints", len(m.eps))
	close(m.ex)
}

func (m *EndpointManager) reloadEndpoints() {
	log.Printf("Endpoints Reloading")

	var done sync.WaitGroup
	for _, ep := range m.eps {
		done.Add(1)
		go func(ep *Endpoint) {
			if err := m.restartEndpoint(ep.Config); err != nil {
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
func (m *EndpointManager) restartEndpoint(cfg config.EndpointConfig) error {
	if old, ok := m.eps[cfg.Name]; ok {
		old.Stop()
		delete(m.eps, cfg.Name)
	}
	ch, err := m.conn.Channel()
	if err != nil {
		return err
	}
	ep, err := New(ch, cfg)
	if err != nil {
		return err
	}
	m.eps[cfg.Name] = ep
	return nil
}

func (m *EndpointManager) stopAllEndpoints() error {
	var done sync.WaitGroup
	for _, ep := range m.eps {
		done.Add(1)
		go func(ep *Endpoint) {
			if err := m.stopEndpoint(ep.Config.Name); err != nil {
				// @todo errors here might put things in a bad state
				log.Println(err)
			}
			done.Done()
		}(ep)
	}
	done.Wait()
	return nil
}

// Stop endpoint and remove from index
func (m *EndpointManager) stopEndpoint(name string) error {
	old, ok := m.eps[name]
	if !ok {
		return fmt.Errorf("%s not present, can't stop", name)
	}
	old.Stop()
	delete(m.eps, name)
	return nil
}
