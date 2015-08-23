package ep

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/streadway/amqp"
)

//func New(ap clb.AddressProvider, cfg Config) *Manager {
func NewManager(cfgMgr *ConfigManager) *Manager {

	return &Manager{
		//		Ap:     ap,
		eps:    make(map[string]*Endpoint),
		epErrs: make(chan EpError),
		cfgMgr: cfgMgr,
	}
}

type Manager struct {
	//	Ap     clb.AddressProvider
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
				ch, err := m.conn.Channel()
				if err != nil {
					return err
				}
				ep := New(ch, cfgU.Config, m.epErrs)
				if err := ep.Start(); err != nil {
					return err
				}
				m.eps[cfgU.Config.Name] = ep
			case ConfigUpdateDelete:
				if m.eps[cfgU.Config.Name] == nil {
					log.Printf("%s not present, can't stop", cfgU.Config.Name)
				}
				m.eps[cfgU.Config.Name].Stop()
				delete(m.eps, cfgU.Config.Name)

			}
		}
	}
	return nil
}
func (m *Manager) Reload() {
	log.Printf("Reloading Endpoints")

	for name, ep := range m.eps {
		ch, err := m.conn.Channel()
		if err != nil {
			// @todo Handle Me!
			log.Println(err)
			continue
		}

		if err := ep.Stop(); err != nil {
			// @todo Handle Me!
			log.Println(err)
			continue
		}

		newEp := New(ch, ep.Config, m.epErrs)
		if err := newEp.Start(); err != nil {
			// @todo Handle Me!
			log.Println(err)
			continue
		}
		m.eps[name] = newEp
	}
	log.Printf("Reloaded Endpoints")
}

func (m *Manager) Stop() {
	log.Printf("Stopping %d Endpoints", len(m.eps))
	defer m.conn.Close()

	exitErrs := make(chan error)
	for _, ep := range m.eps {
		go func(ep *Endpoint) {
			exitErrs <- ep.Stop()
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
