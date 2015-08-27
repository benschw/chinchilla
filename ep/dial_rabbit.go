package ep

import (
	"github.com/benschw/chinchilla/config"
	"github.com/streadway/amqp"
)

func DialRabbit(ap config.RabbitAddressProvider) (*amqp.Connection, chan *amqp.Error, error) {
	add, err := ap.GetAddress()
	if err != nil {
		return nil, nil, err
	}
	conn, err := amqp.Dial(add.String())
	if err != nil {
		return nil, nil, err
	}
	connErr := conn.NotifyClose(make(chan *amqp.Error))
	return conn, connErr, nil
}
