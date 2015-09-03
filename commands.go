package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/benschw/chinchilla/config"
	"github.com/benschw/chinchilla/ep"
	"github.com/benschw/chinchilla/ep/queue"
	"github.com/benschw/dns-clb-go/clb"
	"github.com/hashicorp/consul/api"
	"github.com/xordataexchange/crypt/encoding/secconf"
)

func DoCryptUtil(cmd string, in string, keyring string, secretKeyring string) (string, error) {
	switch cmd {
	case "encrypt":
		if keyring == "" {
			return "", fmt.Errorf("-keyring requred to encrypt")
		}
		return encrypt(keyring, in)
	case "decrypt":
		if keyring == "" {
			return "", fmt.Errorf("-secret-keyring requred to decrypt")
		}
		return decrypt(secretKeyring, in)
	}
	return "", fmt.Errorf("Invalid Subcommand: %s", cmd)
}
func encrypt(kPath string, in string) (string, error) {
	kr, err := os.Open(kPath)
	if err != nil {
		return "", err
	}
	bytes, err := secconf.Encode([]byte(in), kr)
	return string(bytes[:]), nil
}
func decrypt(sKPath string, encrypted string) (string, error) {
	kr, err := os.Open(sKPath)
	if err != nil {
		return "", err
	}
	bytes, err := secconf.Decode([]byte(encrypted), kr)
	return string(bytes[:]), nil
}
func StartDaemon(configPath string, sKPath string) error {

	var kr []byte
	if sKPath != "" {
		kRing, err := os.Open(sKPath)
		if err != nil {
			return err
		}
		bytes, err := ioutil.ReadAll(kRing)
		if err != nil {
			return err
		}
		kr = bytes
	}
	// lb := clb.New()
	lb := clb.NewClb("127.0.0.1", "8600", clb.RoundRobin)

	var ap config.RabbitAddressProvider
	var epp config.EndpointsProvider

	if configPath != "" {
		repo := &config.YamlRepo{Kr: kr, Lb: lb, Path: configPath}

		ap = repo
		epp = repo
	} else {
		client, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		repo := &config.ConsulRepo{Kr: kr, Lb: lb, Client: client}
		ap = repo
		epp = repo
	}

	qReg := ep.NewQueueRegistry()
	qReg.Add(qReg.DefaultKey, &queue.Queue{C: &queue.DefaultWorker{}, D: &queue.DefaultDeliverer{}})

	svc := ep.NewApp(ap, epp, qReg)
	return svc.Run()
}
