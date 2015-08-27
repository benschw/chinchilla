package config

// Satisfy the interface, but just pass through static data
type StaticRepo struct {
	Address   RabbitAddress
	Endpoints []EndpointConfig
}

func (c *StaticRepo) GetEndpoints() ([]EndpointConfig, error) {
	return c.Endpoints, nil
}

func (a *StaticRepo) GetAddress() (RabbitAddress, error) {
	return a.Address, nil
}
