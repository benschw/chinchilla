package ep

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/streadway/amqp"
)

func NewManager(ap RabbitAddressProvider, cfgMgr *ConfigManager) *Manager {
	return &Manager{
		ap:        ap,
		eps:       make(map[string]*Endpoint),
		cfgMgr:    cfgMgr,
		ttl:       5,
		connRetry: 2,
	}
}

type Manager struct {
	ap        RabbitAddressProvider
	conn      *amqp.Connection
	connErr   chan *amqp.Error
	eps       map[string]*Endpoint
	cfgMgr    *ConfigManager
	ttl       int
	connRetry int
}

func (m *Manager) connect() error {
	add, err := m.ap.Get()
	if err != nil {
		return err
	}
	conn, err := amqp.Dial(add.String())
	if err != nil {
		return err
	}
	m.conn = conn
	m.connErr = m.conn.NotifyClose(make(chan *amqp.Error))
	return nil
}

func (m *Manager) Run() error {
	log.Println("Starting...")
	go m.cfgMgr.Manage(m.ttl)

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
		case cfgU := <-m.cfgMgr.Updates:
			switch cfgU.T {
			case ConfigUpdateUpdate:
				if err := m.restartEndpoint(cfgU.Config); err != nil {
					log.Printf("%s: problem starting/reloading: %s", cfgU.Config.Name, err)
				}
			case ConfigUpdateDelete:
				if err := m.stopEndpoint(cfgU.Name); err != nil {
					log.Printf("%s: problem stopping: %s", cfgU.Name, err)
				}
			}
		}
	}
	return nil
}

func (m *Manager) Reload() {
	log.Printf("Endpoint Manager Reloading")
	if err := m.connect(); err != nil {
		// just log out the problem and don't try to recover
		log.Printf("Couldn't reconnect: %s", err)
		return
	}
	m.reloadEndpoints()
	log.Printf("Endpoint Manager Reloaded")
}
func (m *Manager) Stop() {
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

func (m *Manager) reloadEndpoints() {
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
func (m *Manager) restartEndpoint(cfg EndpointConfig) error {
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
func (m *Manager) stopEndpoint(name string) error {
	old, ok := m.eps[name]
	if !ok {
		return fmt.Errorf("%s not present, can't stop", name)
	}
	old.Stop()
	delete(m.eps, name)
	return nil
}
