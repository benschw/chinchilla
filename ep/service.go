package ep

import (
	"github.com/benschw/dns-clb-go/clb"
	"github.com/streadway/amqp"
)

func New(ap clb.AddressProvider, cfg Config) *Service {
	return &Service{
		Ap:     ap,
		Config: cfg,
	}
}

type Service struct {
	Ap     clb.AddressProvider
	Config Config
}

func (s *Service) Run() error {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return err
	}
	defer conn.Close()

	forever := make(chan bool)

	eps := make([]*Endpoint, 0)

	for _, cfg := range s.Config.Endpoints {
		ep := NewEndpoint(conn, cfg)
		go ep.Run()
		eps = append(eps, ep)
	}

	<-forever
	return nil
}
