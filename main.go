package main

import (
	"flag"
	"log"
	"log/syslog"
	"os"

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
	var ap ep.RabbitAddressProvider
	eps := make([]ep.ConfigProvider, 0)

	if *configPath != "" {
		ap = &ep.YamlRabbitAddressProvider{Path: *configPath}
		eps = append(eps, &ep.YamlConfigProvider{Path: *configPath})
	} else {
		client, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		ap = &ep.ConsulRabbitAddressProvider{Client: client}
		eps = append(eps, &ep.ConsulConfigProvider{Client: client})
	}

	cfgMgr := ep.NewConfigManager(eps)

	svc := ep.NewManager(ap, cfgMgr)
	if err := svc.Run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}

}
