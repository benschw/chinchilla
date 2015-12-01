package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/benschw/chinchilla/config"
	"github.com/benschw/chinchilla/ep"
	_ "github.com/benschw/chinchilla/queue"
	_ "github.com/benschw/chinchilla/queue/repeater"
	"github.com/benschw/srv-lb/lb"
	"github.com/hashicorp/consul/api"
	"github.com/xordataexchange/crypt/encoding/secconf"
)

func Encrypt(kPath string, in string) (string, error) {
	kr, err := os.Open(kPath)
	if err != nil {
		return "", err
	}
	bytes, err := secconf.Encode([]byte(in), kr)
	return string(bytes[:]), nil
}
func Decrypt(sKPath string, encrypted string) (string, error) {
	kr, err := os.Open(sKPath)
	if err != nil {
		return "", err
	}
	bytes, err := secconf.Decode([]byte(encrypted), kr)
	return string(bytes[:]), nil
}
func StartDaemon(configPath string, consulPath string, sKPath string) error {

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
	lb := lb.NewGeneric(lb.DefaultConfig())

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
		repo := &config.ConsulRepo{ConsulPath: consulPath, Kr: kr, Lb: lb, Client: client}
		ap = repo
		epp = repo
	}

	svc := ep.NewApp(ap, epp)
	return svc.Run()
}
