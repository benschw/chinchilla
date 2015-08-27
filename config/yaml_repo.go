package config

import "github.com/benschw/opin-go/config"

// Load Config from a yaml file on disk
type YamlRepo struct {
	Path string
}

func (r *YamlRepo) GetEndpoints() ([]EndpointConfig, error) {
	var cfg Config

	if err := config.Bind(r.Path, &cfg); err != nil {
		return make([]EndpointConfig, 0), err
	}

	return cfg.Endpoints, nil
}

func (r *YamlRepo) GetAddress() (RabbitAddress, error) {
	var cfg Config

	if err := config.Bind(r.Path, &cfg); err != nil {
		return RabbitAddress{}, err
	}
	return connectionConfigToAddress(cfg.Connection)
}
