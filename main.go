package main

import (
	"flag"
	"log"
	"log/syslog"
	"os"

	"github.com/benschw/chinchilla/ep"
	"github.com/benschw/opin-go/config"
)

func main() {
	useSyslog := flag.Bool("syslog", false, "log to syslog")
	flag.Parse()

	if *useSyslog {
		logwriter, err := syslog.New(syslog.LOG_NOTICE, "todo")
		if err == nil {
			log.SetOutput(logwriter)
		}
	}
	var cfg ep.Config

	if err := config.Bind("./config.yaml", &cfg); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	//ap := clb.NewAddressProvider("rabbit.service.consul")
	//	ap := &clb.StaticAddressProvider{Address: dns.Address{
	//		Address: "localhost",
	//		Port:    5672,
	//	}}
	svc := ep.New(cfg)

	if err := svc.Run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
