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
		eps:    make([]*Endpoint, 0),
	}
}

type Service struct {
	//	Ap     clb.AddressProvider
	Config Config
	conn   *amqp.Connection
	eps    []*Endpoint
}

func (s *Service) Run() error {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return err
	}
	s.conn = conn

	epErrs := make(chan EpError)

	for _, cfg := range s.Config.Endpoints {
		ch, err := s.conn.Channel()
		if err != nil {
			return err
		}
		ep := NewEndpoint(ch, cfg, epErrs)
		if err := ep.Start(); err != nil {
			return err
		}
		s.eps = append(s.eps, ep)
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
				return nil
			case syscall.SIGHUP:
				s.Reload()
			}
		case err := <-epErrs:
			log.Printf("Handling screwed up %s Endpoint: %s", err.Name, err.Err)
		}
	}
	return nil
}
func (s *Service) Reload() {
	log.Printf("Reloading Endpoints")

	for _, ep := range s.eps {
		ch, err := s.conn.Channel()
		if err != nil {
			// @todo Handle Me!
			log.Println(err)
			continue
		}
		err = ep.Reload(ch, ep.Config)
		if err != nil {
			// @todo Handle Me!
			log.Println(err)
			continue
		}
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
