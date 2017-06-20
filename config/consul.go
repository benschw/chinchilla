package config

import (
	"fmt"

	consul "github.com/hashicorp/consul/api"
)

func NewConsulClient() (*ConsulClient, error) {
	config := consul.DefaultConfig()
	c, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &ConsulClient{consul: c}, nil
}

type ConsulClient struct {
	consul *consul.Client
}

func (c *ConsulClient) Service(service, tag string) ([]*consul.ServiceEntry, error) {
	passingOnly := true
	addrs, _, err := c.consul.Health().Service(service, tag, passingOnly, nil)
	if len(addrs) == 0 && err == nil {
		return nil, fmt.Errorf("service ( %s ) was not found", service)
	}
	if err != nil {
		return nil, err
	}
	return addrs, nil
}
