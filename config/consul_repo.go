package config

import (
	"log"

	"github.com/benschw/dns-clb-go/clb"
	"github.com/hashicorp/consul/api"
	"gopkg.in/yaml.v2"
)

// Load config from Consul
type ConsulRepo struct {
	Lb     clb.LoadBalancer
	Client *api.Client
}

func (r *ConsulRepo) GetEndpoints() ([]EndpointConfig, error) {
	arr := make([]EndpointConfig, 0)

	root := "chinchilla/endpoints/"
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
			log.Println(err)
			continue
		}
		arr = append(arr, *epCfg)
	}
	return arr, nil
}

func (r *ConsulRepo) GetAddress() (RabbitAddress, error) {
	kv := r.Client.KV()

	p, _, err := kv.Get("chinchilla/connection.yaml", nil)
	connCfg := &ConnectionConfig{}

	if err = yaml.Unmarshal(p.Value, connCfg); err != nil {
		return RabbitAddress{}, err
	}

	return connectionConfigToAddress(*connCfg, r.Lb)
}
