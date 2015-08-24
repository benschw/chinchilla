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
	flag.Parse()

	if *useSyslog {
		logwriter, err := syslog.New(syslog.LOG_NOTICE, "chinchilla")
		if err == nil {
			log.SetOutput(logwriter)
		}
	}

	ap := &ep.StaticRabbitAddressProvider{
		Address: ep.RabbitAddress{
			User:     "guest",
			Password: "guest",
			Host:     "localhost",
			Port:     5672,
		},
	}
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	cfgMgr := ep.NewConfigManager([]ep.ConfigProvider{
		//&ep.YamlConfigProvider{Path: "./config.yaml"},
		&ep.ConsulConfigProvider{Client: client},
	})

	svc := ep.NewManager(ap, cfgMgr)
	if err := svc.Run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}

}
