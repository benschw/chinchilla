package ep

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/streadway/amqp"
)

//func New(ap clb.AddressProvider, cfg Config) *Service {
func New(cfg Config) *Service {

	return &Service{
		//		Ap:     ap,
		Config: cfg,
		eps:    make(map[string]*Endpoint),
		epErrs: make(chan EpError),
	}
}

type Service struct {
	//	Ap     clb.AddressProvider
	Config Config
	conn   *amqp.Connection
	eps    map[string]*Endpoint
	epErrs chan EpError
}

func (s *Service) Run() error {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return err
	}
	s.conn = conn

	for _, cfg := range s.Config.Endpoints {
		ch, err := s.conn.Channel()
		if err != nil {
			return err
		}
		ep := NewEndpoint(ch, cfg, s.epErrs)
		if err := ep.Start(); err != nil {
			return err
		}
		s.eps[cfg.Name] = ep
	}

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
				s.Stop()
				return nil
			case syscall.SIGHUP:
				s.Reload()
			}
		case err := <-s.epErrs:
			delete(s.eps, err.Name)
			log.Printf("%s endpoint just errored out: %s", err.Name, err.Err)
		}
	}
	return nil
}
func (s *Service) Reload() {
	log.Printf("Reloading Endpoints")

	for name, ep := range s.eps {
		ch, err := s.conn.Channel()
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

		newEp := NewEndpoint(ch, ep.Config, s.epErrs)
		if err := newEp.Start(); err != nil {
			// @todo Handle Me!
			log.Println(err)
			continue
		}
		s.eps[name] = newEp
	}
	log.Printf("Reloaded Endpoints")
}

func (s *Service) Stop() {
	log.Printf("Stopping %d Endpoints", len(s.eps))
	defer s.conn.Close()

	exitErrs := make(chan error)
	for _, ep := range s.eps {
		go func(ep *Endpoint) {
			exitErrs <- ep.Stop()
		}(ep)
	}

	for i := 0; i < len(s.eps); i++ {
		err := <-exitErrs
		if err != nil {
			// store these and handle separately? can't just stop processing though
			log.Println(err)
		}
	}

	log.Printf("All Endpoints Stopped")
}
