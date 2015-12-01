package repeater

import "github.com/benschw/chinchilla/config"

type ConnectionAddressProvider struct {
	add config.RabbitAddress
}

func (c *ConnectionAddressProvider) GetAddress() (config.RabbitAddress, error) {
	return c.add, nil
}
