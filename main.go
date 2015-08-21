package main

import (
	"flag"
	"log"
	"log/syslog"
	"os"
	"os/signal"
	"syscall"

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
	if err := svc.Start(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer svc.Stop()

	// impl control flow with signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGHUP)

	for {
		sig := <-sigCh
		switch sig {
		case os.Interrupt:
			fallthrough
		case syscall.SIGTERM:
			return
		case syscall.SIGHUP:
			svc.Reload()
		}
	}
}
