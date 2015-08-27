package config

import (
	"log"
	"time"
)

type ConfigUpdateType int

const (
	EndpointUpdate ConfigUpdateType = iota
	EndpointDelete ConfigUpdateType = iota
)

// Sent over Update chan for Endpoint Mgr to operate on
type ConfigUpdate struct {
	T      ConfigUpdateType
	Name   string
	Config EndpointConfig
}

func NewWatcher(p EndpointsProvider) *ConfigWatcher {
	return &ConfigWatcher{
		Provider: p,
		Updates:  make(chan ConfigUpdate, 5),
		cache:    make(map[string]EndpointConfig),
	}
}

// Coordinates config providers and delivers ep updates over the Updates chan
type ConfigWatcher struct {
	Provider EndpointsProvider
	Updates  chan ConfigUpdate
	cache    map[string]EndpointConfig
}

// Poll EndpointsProvider looking for EndpointConfig updates
func (c *ConfigWatcher) Watch(ttl int) {
	for {
		if err := c.processProviders(); err != nil {
			log.Println("Problem loading config, keeping old configuration")
		}
		time.Sleep(time.Duration(ttl) * time.Second)
	}
}

func (c *ConfigWatcher) processProviders() error {
	epCfgs := make(map[string]EndpointConfig)

	// capture all EndpointConfigs, return/abort if problems
	eps, err := c.Provider.GetEndpoints()
	if err != nil {
		return err
	}

	// index endpoints
	for _, ec := range eps {
		epCfgs[ec.Name] = ec
	}

	// notify of missing configs, remove from cache
	for n, _ := range c.cache {
		if _, ok := epCfgs[n]; !ok {
			c.Updates <- ConfigUpdate{
				T:    EndpointDelete,
				Name: n,
			}
			delete(c.cache, n)
		}
	}

	// notify of updated configs, update cache
	for n, cfg := range epCfgs {

		if !cfg.Equals(c.cache[n]) {
			c.Updates <- ConfigUpdate{
				T:      EndpointUpdate,
				Config: cfg,
			}
			c.cache[n] = cfg
		}
	}

	return nil
}
