package ep

import (
	"log"
	"time"

	"github.com/benschw/chinchilla/config"
	"github.com/streadway/amqp"
)

// qReg := NewQueueRegistry()
// qReg.Add(qReg.DefaultWorker, &queue.Queue{C: *queue.MsgConsumer{}, D: *queue.MsgDeliverer{}})
func NewApp(ap config.RabbitAddressProvider, epp config.EndpointsProvider) *EndpointApp {
	return &EndpointApp{
		ap:        ap,
		epp:       epp,
		eps:       NewEndpointMgr(),
		ttl:       5,
		connRetry: 2,
		ex:        make(chan struct{}, 1),
	}
}

type EndpointApp struct {
	ap        config.RabbitAddressProvider
	epp       config.EndpointsProvider
	conn      *amqp.Connection
	connErr   chan *amqp.Error
	eps       *EndpointMgr
	ttl       int
	connRetry int
	ex        chan struct{}
}

func (m *EndpointApp) connect() error {
	conn, connErr, err := DialRabbit(m.ap)
	if err != nil {
		log.Println("Problem connecting to Rabbitmq")
		return err
	}
	m.conn = conn
	m.connErr = connErr
	return nil
}

func (m *EndpointApp) Run() error {
	log.Println("Starting...")

	if err := m.connect(); err != nil {
		return err
	}
	defer m.conn.Close()

	cfgWatcher := config.NewWatcher(m.epp, m.ttl)
	defer cfgWatcher.Stop()

	sigWatcher := NewSignalWatcher()
	defer sigWatcher.Stop()

	// main control flow
	for {
		select {
		// Handle incoming config updates
		case cfgU := <-cfgWatcher.Updates:
			switch cfgU.T {
			case config.EndpointUpdate:
				if err := m.eps.RestartEndpoint(m.conn, cfgU.Config); err != nil {
					log.Printf("%s: problem starting/reloading: %s", cfgU.Config.Name, err)
				}
			case config.EndpointDelete:
				if err := m.eps.StopEndpoint(cfgU.Name); err != nil {
					log.Printf("%s: problem stopping: %s", cfgU.Name, err)
				}
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
				break
			}
			m.eps.RestartAllEndpoints(m.conn)

		// If a signal is caught, either shutdown or reload gracefully
		case t := <-sigWatcher.T:
			switch t {
			case TriggerStop:
				m.Stop()
			case TriggerReload:
				m.Reload()
			}

		// `Stop` func signals this chan to break out of the main loop
		case <-m.ex:
			m.eps.StopAllEndpoints()
			return nil
		}
	}
}

func (m *EndpointApp) Reload() {
	log.Printf("Endpoint Manager Reloading")
	if err := m.connect(); err != nil {
		// just log out the problem and don't try to recover
		log.Printf("Couldn't reconnect: %s", err)
		return
	}
	m.eps.RestartAllEndpoints(m.conn)
	log.Printf("Endpoint Manager Reloaded")
}
func (m *EndpointApp) Stop() {
	close(m.ex)
}
