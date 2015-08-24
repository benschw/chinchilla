package main

import (
	"flag"
	"log"
	"log/syslog"
	"os"

	"github.com/benschw/chinchilla/ep"
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

	cfgMgr := ep.NewConfigManager([]ep.ConfigProvider{
		&ep.YamlConfigProvider{Path: "./config.yaml"},
	})

	svc := ep.NewManager(ap, cfgMgr)
	if err := svc.Run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}

}
