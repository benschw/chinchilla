package ep

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/streadway/amqp"
)

func NewManager(cfgMgr *ConfigManager) *Manager {
	return &Manager{
		eps:    make(map[string]*Endpoint),
		epErrs: make(chan EpError),
		cfgMgr: cfgMgr,
	}
}

type Manager struct {
	conn   *amqp.Connection
	eps    map[string]*Endpoint
	epErrs chan EpError
	cfgMgr *ConfigManager
}

func (m *Manager) Run() error {
	go m.cfgMgr.Manage(5)

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return err
	}
	m.conn = conn

	log.Println("starting up")
	// control flow with signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGHUP)

	for {
		select {
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
		case err := <-m.epErrs:
			delete(m.eps, err.Name)
			log.Printf("%s endpoint just errored out: %s", err.Name, err.Err)
		case cfgU := <-m.cfgMgr.Updates:
			switch cfgU.T {
			case ConfigUpdateUpdate:
				if err := m.startEndpoint(cfgU.Config); err != nil {
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
	log.Printf("Reloading Endpoints")

	for _, ep := range m.eps {
		if err := m.startEndpoint(ep.Config); err != nil {
			log.Println(err)
			continue
		}
	}
	log.Printf("Reloaded Endpoints")
}

func (m *Manager) Stop() {
	log.Printf("Stopping %d Endpoints", len(m.eps))
	defer m.conn.Close()
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in f", r)
		}
	}()
	exitErrs := make(chan error)
	for _, ep := range m.eps {
		go func(ep *Endpoint) {
			ep.Stop()
			exitErrs <- nil
		}(ep)
	}

	for i := 0; i < len(m.eps); i++ {
		err := <-exitErrs
		if err != nil {
			// store these and handle separately? can't just stop processing though
			log.Println(err)
		}
	}

	log.Printf("All Endpoints Stopped")
}
func (m *Manager) startEndpoint(cfg EndpointConfig) error {
	if old, ok := m.eps[cfg.Name]; ok {
		old.Stop()
		delete(m.eps, cfg.Name)
	}
	ch, err := m.conn.Channel()
	if err != nil {
		return err
	}
	ep := New(ch, cfg, m.epErrs)
	if err := ep.Start(); err != nil {
		return err
	}
	m.eps[cfg.Name] = ep
	return nil
}
func (m *Manager) stopEndpoint(name string) error {
	old, ok := m.eps[name]
	if !ok {
		return fmt.Errorf("%s not present, can't stop", name)
	}
	old.Stop()
	delete(m.eps, name)
	return nil
}
