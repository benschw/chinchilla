package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/benschw/chinchilla/config"
	"github.com/benschw/chinchilla/ep"
	_ "github.com/benschw/chinchilla/queue"
	"github.com/benschw/srv-lb/dns"
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
func StartDaemon(configPath string, conConfigPath string, consulPath string, sKPath string) error {

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
	lb.RegisterStrategy(FirstStrategy, NewFirstStrategy)

	lbCfg := lb.DefaultConfig()
	lbCfg.Strategy = FirstStrategy

	lb := lb.NewGeneric(lbCfg)

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

	if conConfigPath != "" {
		ap = &config.YamlRepo{Kr: kr, Lb: lb, Path: conConfigPath}
	}

	svc := ep.NewApp(ap, epp)
	return svc.Run()
}

const FirstStrategy lb.StrategyType = "first"

func NewFirstStrategy(lib dns.Lookup) lb.GenericLoadBalancer {
	return &FirstClb{lib}
}

type FirstClb struct {
	dnsLib dns.Lookup
}

func (lb *FirstClb) Next(name string) (dns.Address, error) {
	var add dns.Address

	srvs, err := lb.dnsLib.LookupSRV(name)
	if err != nil {
		return add, err
	}

	ip, err := lb.dnsLib.LookupA(srvs[0].Target)
	if err != nil {
		return add, err
	}

	return dns.Address{Address: ip, Port: srvs[0].Port}, nil
}
