package main

import (
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"os"
)

func init() {
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

	useSyslog := flag.Bool("syslog", false, "log to syslog")
	configPath := flag.String("config", "", "path to yaml config. omit to use consul")
	keyring := flag.String("keyring", "", "path to armored public keyring")
	secretKeyring := flag.String("secret-keyring", "", "path to armored secret keyring")
	flag.Parse()

	if *useSyslog {
		logwriter, err := syslog.New(syslog.LOG_NOTICE, "chinchilla")
		if err == nil {
			log.SetOutput(logwriter)
		}
	}
	if flag.NArg() == 0 {
		// If no subcommands, run daemon
		if err := StartDaemon(*configPath, *secretKeyring); err != nil {
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

		out, err := DoCryptUtil(cmd, in, *keyring, *secretKeyring)
		if err != nil {
			fmt.Println(err)
			flag.Usage()
			os.Exit(1)
		}
		fmt.Println(out)
	}
}
