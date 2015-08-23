package ep

import (
	"log"
	"time"
)

type ConfigUpdateType int

const (
	ConfigUpdateUpdate ConfigUpdateType = iota
	ConfigUpdateDelete ConfigUpdateType = iota
)

// Sent over Update chan for Endpoint Mgr to operate on
type ConfigUpdate struct {
	T      ConfigUpdateType
	Name   string
	Config EndpointConfig
}

func NewConfigManager(ps []ConfigProvider) *ConfigManager {
	return &ConfigManager{
		Providers: ps,
		Updates:   make(chan ConfigUpdate, 5),
		cache:     make(map[string]EndpointConfig),
	}
}

// Coordinates config providers and delivers ep updates over the Updates chan
type ConfigManager struct {
	Providers []ConfigProvider
	Updates   chan ConfigUpdate
	cache     map[string]EndpointConfig
}

func (c *ConfigManager) Manage(ttl int) {

	for {
		if err := c.processProviders(); err != nil {
			log.Println("Problem loading config, keeping old configuration")
		}
		time.Sleep(time.Duration(ttl) * time.Second)
	}

}
func (c *ConfigManager) processProviders() error {
	epCfgs := make(map[string]EndpointConfig)

	// capture all EndpointConfigs, return/abort if problems
	for _, p := range c.Providers {
		cfg, err := p.GetConfig()
		if err != nil {
			return err
		}

		if cfg.Endpoints != nil {
			for _, ec := range cfg.Endpoints {
				epCfgs[ec.Name] = ec
			}
		}
	}

	// notify of missing configs, remove from cache
	for n, _ := range c.cache {
		if _, ok := epCfgs[n]; !ok {
			c.Updates <- ConfigUpdate{
				T:    ConfigUpdateDelete,
				Name: n,
			}
			delete(c.cache, n)
		}
	}

	// notify of updated configs, update cache
	for n, cfg := range epCfgs {

		if !cfg.Equals(c.cache[n]) {
			c.Updates <- ConfigUpdate{
				T:      ConfigUpdateUpdate,
				Config: cfg,
			}
			c.cache[n] = cfg
		}
	}

	return nil
}
