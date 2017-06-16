package config

import (
	"log"
	"os"
	"strconv"

	"github.com/benschw/srv-lb/lb"
)

func getRabbitmqPassword(l lb.GenericLoadBalancer, secretsPath string) (string, error) {
	if _, found := os.LookupEnv("VAULT_APPROLE_PATH"); found {
		log.Printf("using rabbitmq password from vault")
		return getRabbitmqPasswordFromVault(l, secretsPath)
	}
	log.Printf("using rabbitmq password from environment variable")

	return os.Getenv("RABBITMQ_PASSWORD"), nil
}

func NewEnvRabbitAp(l lb.GenericLoadBalancer, secretsPath string) (*EnvRabbitAp, error) {
	var err error
	var port int64 = 5672
	portStr, found := os.LookupEnv("RABBITMQ_PORT")
	if found {
		if port, err = strconv.ParseInt(portStr, 10, 16); err != nil {
			return nil, err
		}
	}

	password, err := getRabbitmqPassword(l, secretsPath)
	if err != nil {
		return nil, err
	}

	return &EnvRabbitAp{
		Conc: ConnectionConfig{
			User:        os.Getenv("RABBITMQ_USER"),
			Password:    password,
			ServiceName: os.Getenv("RABBITMQ_SERVICENAME"),
			Host:        os.Getenv("RABBITMQ_HOST"),
			Port:        uint16(port),
			VHost:       os.Getenv("RABBITMQ_VHOST"),
		},
		Lb: l,
	}, nil

}

type EnvRabbitAp struct {
	Conc ConnectionConfig
	Lb   lb.GenericLoadBalancer
}

func (r *EnvRabbitAp) GetAddress() (RabbitAddress, error) {
	add := RabbitAddress{
		User:     r.Conc.User,
		Password: r.Conc.Password,
		Host:     r.Conc.Host,
		Port:     r.Conc.Port,
		VHost:    r.Conc.VHost,
	}

	if r.Conc.ServiceName != "" {
		a, err := r.Lb.Next("rabbitmq.service.consul")
		if err != nil {
			return add, err
		}
		add.Host = a.Address
		add.Port = a.Port
	}
	log.Printf("rabbitmq address: %s:%d ", add.Host, add.Port)
	return add, nil
}
