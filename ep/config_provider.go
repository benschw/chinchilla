package ep

import (
	"log"

	"gopkg.in/yaml.v2"

	"github.com/benschw/opin-go/config"
	"github.com/hashicorp/consul/api"
)

type ConfigProvider interface {
	GetConfig() ([]EndpointConfig, error)
}

// Load Config from a yaml file on disk
type YamlConfigProvider struct {
	Path string
}

func (c *YamlConfigProvider) GetConfig() ([]EndpointConfig, error) {
	var cfg Config

	if err := config.Bind(c.Path, &cfg); err != nil {
		return make([]EndpointConfig, 0), err
	}

	return cfg.Endpoints, nil
}

// Satisfy the interface, but just pass through static data
type StaticConfigProvider struct {
	Endpoints []EndpointConfig
}

func (c *StaticConfigProvider) GetConfig() ([]EndpointConfig, error) {
	return c.Endpoints, nil
}

// Load config from Consul
// @todo
type ConsulConfigProvider struct {
	Client *api.Client
}

func (c *ConsulConfigProvider) GetConfig() ([]EndpointConfig, error) {
	arr := make([]EndpointConfig, 0)

	root := "chinchilla/endpoints/"
	kv := c.Client.KV()

	results, _, err := kv.List(root, nil)
	if err != nil {
		return arr, err
	}

	for _, p := range results {
		if p.Key == root {
			continue
		}

		epCfg := &EndpointConfig{}

		if err = yaml.Unmarshal(p.Value, epCfg); err != nil {
			log.Println(err)
			continue
		}
		arr = append(arr, *epCfg)
	}
	return arr, nil
}
