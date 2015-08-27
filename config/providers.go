package config

type ConfigProvider interface {
	EndpointsProvider
	RabbitAddressProvider
}

type EndpointsProvider interface {
	GetEndpoints() ([]EndpointConfig, error)
}

type RabbitAddressProvider interface {
	GetAddress() (RabbitAddress, error)
}
