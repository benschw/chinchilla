package config

import (
	"log"

	"github.com/benschw/srv-lb/lb"
	"github.com/hashicorp/consul/api"
	"gopkg.in/yaml.v2"
)

// Load config from Consul
type ConsulRepo struct {
	Lb         lb.GenericLoadBalancer
	Client     *api.Client
	ConsulPath string
}

func (r *ConsulRepo) GetEndpoints() ([]EndpointConfig, error) {
	arr := make([]EndpointConfig, 0)

	root := r.ConsulPath + "/endpoints/"
	kv := r.Client.KV()

	results, _, err := kv.List(root, nil)
	if err != nil {
		return arr, err
	}

	for _, p := range results {
		if p.Key == root {
			continue
		}

		epCfg := &EndpointConfig{Lb: r.Lb}

		if err = yaml.Unmarshal(p.Value, epCfg); err != nil {
			log.Printf("Error Unmarshaling EP Config: %s", err)
			continue
		}
		arr = append(arr, *epCfg)
	}
	return arr, nil
}
