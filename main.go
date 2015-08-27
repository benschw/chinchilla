package main

import (
	"flag"
	"log"
	"log/syslog"
	"os"

	"github.com/benschw/chinchilla/config"
	"github.com/benschw/chinchilla/ep"
	"github.com/hashicorp/consul/api"
)

func main() {
	useSyslog := flag.Bool("syslog", false, "log to syslog")
	configPath := flag.String("config", "", "path to yaml config. omit to use consul")
	flag.Parse()

	if *useSyslog {
		logwriter, err := syslog.New(syslog.LOG_NOTICE, "chinchilla")
		if err == nil {
			log.SetOutput(logwriter)
		}
	}

	var ap config.RabbitAddressProvider
	eps := make([]config.EndpointsProvider, 0)

	if *configPath != "" {
		repo := &config.YamlRepo{Path: *configPath}

		ap = repo
		eps = append(eps, repo)
	} else {
		client, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		repo := &config.ConsulRepo{Client: client}
		ap = repo
		eps = append(eps, repo)
	}

	cfgWatcher := config.NewWatcher(eps)

	svc := ep.NewManager(ap, cfgWatcher)
	if err := svc.Run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}

}
