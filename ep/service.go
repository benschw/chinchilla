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
	}
}

type Service struct {
	//	Ap     clb.AddressProvider
	Config Config
}

func (s *Service) Run() error {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return err
	}
	defer conn.Close()

	eps := make([]*Endpoint, 0)

	for _, cfg := range s.Config.Endpoints {
		ep, err := NewEndpoint(conn, cfg)
		if err != nil {
			return err
		}
		go ep.Run()
		eps = append(eps, ep)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, syscall.SIGTERM)

	for {
		sig := <-sigCh
		switch sig {
		case os.Interrupt:
			log.Printf("Stopping %d Endpoints", len(eps))
			stopAllEndpoints(eps)
			log.Printf("All Endpoints Stopped")
			return nil
		case syscall.SIGTERM:
			log.Printf("Reconfiguring... one day")
		}
	}
	return nil
}

func stopAllEndpoints(eps []*Endpoint) {
	ch := make(chan bool)
	for _, ep := range eps {
		go func(ep *Endpoint) {
			ep.Stop()
			ch <- true
		}(ep)
	}
	for i := 0; i < len(eps); i++ {
		<-ch
	}
}
