package repeater

import (
	"fmt"
	"log"
	"os"

	"github.com/benschw/chinchilla/config"
	"github.com/benschw/chinchilla/ep"
	"github.com/hashicorp/consul/api"
	"github.com/streadway/amqp"
	"gopkg.in/yaml.v2"
)

type ConAPLookup interface {
	GetAddressProvider(string) (config.RabbitAddressProvider, error)
}

type ConsulConAddressProvider struct {
	Root   string
	Client *api.Client
}

func (c *ConsulConAddressProvider) GetAddressProvider(dc string) (config.RabbitAddressProvider, error) {
	root := c.Root + "/repeater/connections/"
	kv := c.Client.KV()

	results, _, err := kv.List(root, nil)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	for _, p := range results {
		if p.Key == root {
			continue
		}
		add := &config.RabbitAddress{}

		if err = yaml.Unmarshal(p.Value, add); err != nil {
			log.Printf("Error Unmarshaling Repeater Connection Config: %s", err)
			continue
		}

		if add.Name == dc {
			ap := &ConnectionAddressProvider{add: *add}
			return ap, nil
		}
	}
	return nil, fmt.Errorf("Connection for dc '%s' not found in consul", dc)
}

type ConnectionProvider struct {
	ap ConAPLookup
}

func NewConnectionProvider(ap ConAPLookup) ConnectionProvider {
	return ConnectionProvider{ap: ap}
}

func (c *ConnectionProvider) GetConnection(dc string) (*amqp.Connection, error) {
	ap, err := c.ap.GetAddressProvider(dc)
	if err != nil {
		return nil, err
	}

	conn, _, err := ep.DialRabbit(ap)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
