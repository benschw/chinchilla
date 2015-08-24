package ep

import (
	"strings"

	"github.com/benschw/opin-go/config"
	"github.com/hashicorp/consul/api"
)

type ConfigProvider interface {
	GetConfig() (Config, error)
}

// Load Config from a yaml file on disk
type YamlConfigProvider struct {
	Path string
}

func (c *YamlConfigProvider) GetConfig() (Config, error) {
	var cfg Config

	err := config.Bind(c.Path, &cfg)
	return cfg, err
}

// Satisfy the interface, but just pass through static data
type StaticConfigProvider struct {
	Config Config
}

func (c *StaticConfigProvider) GetConfig() (Config, error) {
	return c.Config, nil
}

// Load config from Consul
// @todo
type ConsulConfigProvider struct {
	Client *api.Client
}

func (c *ConsulConfigProvider) GetConfig() (Config, error) {
	root := "chinchilla/endpoints/"
	cfg := Config{Endpoints: make([]EndpointConfig, 0)}
	kv := c.Client.KV()
	eps := make(map[string]*EndpointConfig)

	pairs, _, err := kv.List(root, nil)
	if err != nil {
		return cfg, err
	}
	for _, p := range pairs {
		if p.Key == root {
			continue
		}
		rem := p.Key[len(root):]
		parts := strings.Split(rem, "/")
		if len(parts) == 2 {
			_, ok := eps[parts[0]]
			if !ok {
				eps[parts[0]] = &EndpointConfig{}
			}
			ep := eps[parts[0]]
			switch parts[1] {
			case "Name":
				ep.Name = string(p.Value)
			case "ServiceName":
				ep.ServiceName = string(p.Value)
			case "ServiceHost":
				ep.ServiceHost = string(p.Value)
			case "Uri":
				ep.Uri = string(p.Value)
			case "Method":
				ep.Method = string(p.Value)
			case "QueueName":
				ep.QueueName = string(p.Value)
			}
		}
	}
	arr := make([]EndpointConfig, 0)
	for _, ep := range eps {
		arr = append(arr, *ep)
	}
	cfg.Endpoints = arr
	return cfg, nil
}
