package ep

import "github.com/streadway/amqp"

//func New(ap clb.AddressProvider, cfg Config) *Service {
func New(cfg Config) (*Service, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil, err
	}

	return &Service{
		//		Ap:     ap,
		conn:   conn,
		Config: cfg,
		eps:    make([]*Endpoint, 0),
	}, nil
}

type Service struct {
	//	Ap     clb.AddressProvider
	Config Config
	conn   *amqp.Connection
	eps    []*Endpoint
}

func (s *Service) Start() error {

	for _, cfg := range s.Config.Endpoints {
		ch, err := s.conn.Channel()
		if err != nil {
			return err
		}
		ep, err := NewEndpoint(ch, cfg)
		if err != nil {
			return err
		}
		if err := ep.Start(); err != nil {
			return err
		}
		s.eps = append(s.eps, ep)
	}

	return nil
}
func (s *Service) Stop() {
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

}
