package main

import (
	"log"
	"os"

	"github.com/benschw/chinchilla/config"
	"github.com/benschw/chinchilla/ep"
	_ "github.com/benschw/chinchilla/queue"
	"github.com/benschw/srv-lb/lb"
	"github.com/hashicorp/consul/api"
)

func StartDaemon(configPath string, conConfigPath string, consulPath string) error {

	lb := lb.NewGeneric(lb.DefaultConfig())

	var ap config.RabbitAddressProvider
	var epp config.EndpointsProvider

	if configPath != "" {
		repo := &config.YamlRepo{Lb: lb, Path: configPath}

		ap = repo
		epp = repo
	} else {
		client, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		repo := &config.ConsulRepo{ConsulPath: consulPath, Lb: lb, Client: client}
		ap = repo
		epp = repo
	}

	if conConfigPath != "" {
		ap = &config.YamlRepo{Lb: lb, Path: conConfigPath}
	}

	svc := ep.NewApp(ap, epp)
	return svc.Run()
}
