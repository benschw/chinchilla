package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	_ "expvar"
	"github.com/benschw/chinchilla/config"
	"github.com/benschw/chinchilla/ep"
	_ "github.com/benschw/chinchilla/queue"
	"github.com/benschw/srv-lb/lb"
	"github.com/hashicorp/consul/api"
)

var metricsBind = flag.String("metrics", ":8081", "address to bind metrics to")
var configPath = flag.String("config", "", "path to yaml config. omit to use consul")
var consulPath = flag.String("consul-path", "chinchilla", "consul key path to find configuration in")

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [FLAGS]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Additional help: https://github.com/benschw/chinchilla\n")
	}
}

func main() {


	flag.Parse()

	// Start metrics Server
	sock, err := net.Listen("tcp", *metricsBind)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	go func() {
		http.Serve(sock, nil)
	}()

	// Start Chinchilla daemon
	lb := lb.NewGeneric(lb.DefaultConfig())

	var epp config.EndpointsProvider

	if *configPath != "" {
		epp = &config.YamlRepo{Lb: lb, Path: *configPath}

	} else {
		client, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		epp = &config.ConsulRepo{ConsulPath: *consulPath, Lb: lb, Client: client}
	}

	rabbitAp, err := config.NewEnvRabbitAp(lb)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	svc := ep.NewApp(rabbitAp, epp)
	if err = svc.Run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
