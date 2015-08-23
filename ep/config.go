package ep

import (
	"log"
	"time"

	"github.com/benschw/opin-go/config"
)

type Config struct {
	Endpoints []EndpointConfig `json: "endpoints"`
}

type EndpointConfig struct {
	Name        string `json: "name"`
	ServiceHost string `json: "servicehost"`
	ServiceName string `json: "servicename"`
	Uri         string `json: "uri"`
	Method      string `json: "method"`
	QueueName   string `json: "queuename"`
}

type ConfigUpdateType int

const (
	ConfigUpdateUpdate ConfigUpdateType = iota
	ConfigUpdateDelete ConfigUpdateType = iota
)

type ConfigUpdate struct {
	T      ConfigUpdateType
	Config EndpointConfig
}

func NewConfigManager(ps []ConfigProvider) *ConfigManager {
	return &ConfigManager{
		Providers: ps,
		Updates:   make(chan ConfigUpdate),
	}
}

type ConfigManager struct {
	Providers []ConfigProvider
	Updates   chan ConfigUpdate
}

func (c *ConfigManager) Manage(ttl int) {

	for {
		for _, p := range c.Providers {
			cfg, err := p.GetConfig()
			if err != nil {
				// @todo handle this error, cache last working version
				log.Println("Problem loading config")
			}
			if cfg.Endpoints != nil {
				for _, ec := range cfg.Endpoints {
					c.Updates <- ConfigUpdate{
						T:      ConfigUpdateUpdate,
						Config: ec,
					}
				}
			}
		}
		time.Sleep(time.Duration(ttl) * time.Second)
	}

}

type ConfigProvider interface {
	GetConfig() (Config, error)
}

type YamlConfigProvider struct {
	Path string
}

func (c *YamlConfigProvider) GetConfig() (Config, error) {
	var cfg Config

	err := config.Bind(c.Path, &cfg)
	return cfg, err
}
