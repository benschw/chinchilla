package main

import (
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"net"
	"net/http"
	"os"

	_ "expvar"
)

var metricsBind = flag.String("metrics", ":8081", "address to bind metrics to")
var configPath = flag.String("config", "", "path to yaml config. omit to use consul")
var conConfigPath = flag.String("connection-config", "", "path to yaml connection config. use consul for endpoint configs.")
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

	logwriter, err := syslog.New(syslog.LOG_NOTICE, "chinchilla")
	if err == nil {
		log.SetOutput(logwriter)
	}

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
	if err := StartDaemon(*configPath, *conConfigPath, *consulPath); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
