package ep

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/benschw/chinchilla/config"
	"github.com/streadway/amqp"
)

func NewManager(ap config.RabbitAddressProvider, cfgWatcher *config.ConfigWatcher) *EndpointManager {
	return &EndpointManager{
		ap:         ap,
		eps:        make(map[string]*Endpoint),
		cfgWatcher: cfgWatcher,
		ttl:        5,
		connRetry:  2,
	}
}

type EndpointManager struct {
	ap         config.RabbitAddressProvider
	conn       *amqp.Connection
	connErr    chan *amqp.Error
	eps        map[string]*Endpoint
	cfgWatcher *config.ConfigWatcher
	ttl        int
	connRetry  int
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
	go m.cfgWatcher.Watch(m.ttl)

	if err := m.connect(); err != nil {
		return err
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGHUP)

	// main control flow
	for {
		select {

		// If a signal is caught, either shutdown or reload gracefully
		case sig := <-sigCh:
			switch sig {
			case os.Interrupt:
				fallthrough
			case syscall.SIGTERM:
				m.Stop()
				return nil
			case syscall.SIGHUP:
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
		case cfgU := <-m.cfgWatcher.Updates:
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
	defer m.conn.Close()

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

	log.Printf("All Endpoints Stopped")
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
	ep := New(ch, cfg)
	if err := ep.Start(); err != nil {
		return err
	}
	m.eps[cfg.Name] = ep
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
