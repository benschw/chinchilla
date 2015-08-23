package ep

import "github.com/benschw/opin-go/config"

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
