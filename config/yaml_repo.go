package config

import (
	"github.com/benschw/dns-clb-go/clb"
	"github.com/benschw/opin-go/config"
)

// Load Config from a yaml file on disk
type YamlRepo struct {
	Lb   clb.LoadBalancer
	Path string
}

func (r *YamlRepo) GetEndpoints() ([]EndpointConfig, error) {
	var cfg Config

	if err := config.Bind(r.Path, &cfg); err != nil {
		return make([]EndpointConfig, 0), err
	}

	for i, _ := range cfg.Endpoints {
		cfg.Endpoints[i].Lb = r.Lb
	}

	return cfg.Endpoints, nil
}

func (r *YamlRepo) GetAddress() (RabbitAddress, error) {
	var cfg Config

	if err := config.Bind(r.Path, &cfg); err != nil {
		return RabbitAddress{}, err
	}
	return connectionConfigToAddress(cfg.Connection, r.Lb)
}
