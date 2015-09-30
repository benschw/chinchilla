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

	"github.com/benschw/chinchilla/ep"
	"github.com/benschw/chinchilla/queue"
)

var useSyslog = flag.Bool("syslog", false, "log to syslog")
var logPath = flag.String("log-path", "", "path to log file")
var metricsBind = flag.String("metrics", ":8081", "address to bind metrics to")
var configPath = flag.String("config", "", "path to yaml config. omit to use consul")
var consulPath = flag.String("consul-path", "chinchilla", "consul key path to find configuration in")
var keyring = flag.String("keyring", "", "path to armored public keyring")
var secretKeyring = flag.String("secret-keyring", "", "path to armored secret keyring")

var queueReg = ep.NewQueueRegistry()

func init() {
	queueReg.Add(queueReg.DefaultKey, &queue.Queue{C: &queue.DefaultWorker{}, D: &queue.DefaultDeliverer{}})
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [FLAGS] [SUBCOMMAND]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Subcommands:\n")
		fmt.Fprintf(os.Stderr, "  encrypt  encrypt a single value using `-keyring` \n")
		fmt.Fprintf(os.Stderr, "  decrypt  decrypt a single value using `-secret-keyring` \n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  # Encrypt a value with a gpg public keyring\n")
		fmt.Fprintf(os.Stderr, "  %s -keyring .pubring.gpg encrypt \"my secret\"\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Decrypt a value with a gpg private keyring\n")
		fmt.Fprintf(os.Stderr, "  %s -secret-keyring .secring.gpg decrypt \"wcBMA3WVkZiNgGDU\"\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Start the daemon, configured with a yaml file\n")
		fmt.Fprintf(os.Stderr, "  %s -secret-keyring .secring.gpg -config config.yaml\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Start the daemon, configured with Consul\n")
		fmt.Fprintf(os.Stderr, "  %s -secret-keyring .secring.gpg\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Start the daemon expecting plain text rabbitmq credentials in config\n")
		fmt.Fprintf(os.Stderr, "  %s\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Additional help: https://github.com/benschw/chinchilla\n")
	}
}

func main() {

	flag.Parse()

	if *useSyslog {
		logwriter, err := syslog.New(syslog.LOG_NOTICE, "chinchilla")
		if err == nil {
			log.SetOutput(logwriter)
		}
	} else if *logPath != "" {
		file, err := os.OpenFile(*logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err == nil {
			log.SetOutput(file)
		}
	}

	sock, err := net.Listen("tcp", *metricsBind)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	go func() {
		fmt.Println("HTTP now available at port 8123")
		http.Serve(sock, nil)
	}()

	if flag.NArg() == 0 {
		// If no subcommands, run daemon
		if err := StartDaemon(*configPath, *consulPath, *secretKeyring, queueReg); err != nil {
			log.Println(err)
		}
		os.Exit(1)
	} else {
		// Pull subcommand & input string from args
		if flag.NArg() != 2 {
			flag.Usage()
			os.Exit(1)
		}
		cmd := flag.Arg(0)
		in := flag.Arg(1)

		var out string
		var err error
		switch cmd {
		case "encrypt":
			if *keyring == "" {
				err = fmt.Errorf("-keyring requred to encrypt")
			} else {
				out, err = Encrypt(*keyring, in)
			}
		case "decrypt":
			if *secretKeyring == "" {
				err = fmt.Errorf("-secret-keyring requred to decrypt")
			} else {
				out, err = Decrypt(*secretKeyring, in)
			}
		default:
			err = fmt.Errorf("Invalid Subcommand: %s", cmd)
		}

		if err != nil {
			log.Println(err)
			flag.Usage()
			os.Exit(1)
		}
		fmt.Println(out)
	}
}
