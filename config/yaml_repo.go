package config

import (
	"github.com/benschw/opin-go/config"
	"github.com/benschw/srv-lb/lb"
)

// Load Config from a yaml file on disk
type YamlRepo struct {
	Lb   lb.GenericLoadBalancer
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

