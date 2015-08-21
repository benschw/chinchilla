package ep

import (
	"log"

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

func (s *Service) Start() error {
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
		ep := NewEndpoint(ch, cfg)
		if err := ep.Start(); err != nil {
			return err
		}
		s.eps = append(s.eps, ep)
	}

	return nil
}
func (s *Service) Reload() {
	log.Printf("Reconfiguring... one day")
}

func (s *Service) Stop() {
	log.Printf("Stopping %d Endpoints", len(s.eps))
	defer s.conn.Close()

	ch := make(chan bool)
	for _, ep := range s.eps {
		go func(ep *Endpoint) {
			ep.Stop()
			ch <- true
		}(ep)
	}

	for i := 0; i < len(s.eps); i++ {
		<-ch
	}

	log.Printf("All Endpoints Stopped")
}
