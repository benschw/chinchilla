package config

import (
	"github.com/benschw/opin-go/config"
	"github.com/benschw/srv-lb/srvlb"
)

// Load Config from a yaml file on disk
type YamlRepo struct {
	Kr   []byte
	Lb   srvlb.SRVLoadBalancerDriver
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
	return connectionConfigToAddress(r.Kr, cfg.Connection, r.Lb)
}
