package ep

import (
	"fmt"

	"gopkg.in/yaml.v2"

	"github.com/benschw/opin-go/config"
	"github.com/hashicorp/consul/api"
)

type RabbitAddress struct {
	User     string
	Password string
	Host     string
	Port     int
}

func (a *RabbitAddress) String() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/", a.User, a.Password, a.Host, a.Port)
}

type RabbitAddressProvider interface {
	Get() (RabbitAddress, error)
}

// Satisfy the interface, but just pass through static data
type StaticRabbitAddressProvider struct {
	Address RabbitAddress
}

func (a *StaticRabbitAddressProvider) Get() (RabbitAddress, error) {
	return a.Address, nil
}

// Load connection info from yaml file
type YamlRabbitAddressProvider struct {
	Path string
}

func (c *YamlRabbitAddressProvider) Get() (RabbitAddress, error) {
	var cfg Config

	if err := config.Bind(c.Path, &cfg); err != nil {
		return RabbitAddress{}, err
	}
	return connectionConfigToAddress(cfg.Connection)
}

// Load Connection String from Consul
type ConsulRabbitAddressProvider struct {
	Client *api.Client
}

func (c *ConsulRabbitAddressProvider) Get() (RabbitAddress, error) {
	kv := c.Client.KV()

	p, _, err := kv.Get("chinchilla/connection.yaml", nil)
	connCfg := &ConnectionConfig{}

	if err = yaml.Unmarshal(p.Value, connCfg); err != nil {
		return RabbitAddress{}, err
	}

	return connectionConfigToAddress(*connCfg)
}

func connectionConfigToAddress(c ConnectionConfig) (RabbitAddress, error) {
	// @todo discover if ServiceName is set
	return RabbitAddress{
		User:     c.User,
		Password: c.Password,
		Host:     c.Host,
		Port:     c.Port,
	}, nil
}
