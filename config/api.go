package config

import (
	"bytes"
	"fmt"
	"log"
	"reflect"

	"github.com/benschw/srv-lb/srvlb"
	"github.com/xordataexchange/crypt/encoding/secconf"
)

type Config struct {
	Connection ConnectionConfig `json: "connection"`
	Endpoints  []EndpointConfig `json: "endpoints"`
}

type ConnectionConfig struct {
	User        string `json: "user"`
	Password    string `json: "password"`
	ServiceName string `json: "servicename"`
	Host        string `json: "host"`
	Port        uint16 `json: "port"`
	VHost       string `json: "vhost"`
}

type EndpointConfig struct {
	Lb          srvlb.SRVLoadBalancerDriver
	Name        string                      `json: "name"`
	ServiceHost string                      `json: "servicehost"`
	ServiceName string                      `json: "servicename"`
	Uri         string                      `json: "uri"`
	Method      string                      `json: "method"`
	Prefetch    int                         `json: "prefetch"`
	QueueType   string                      `json: "queuetype"`
	QueueConfig map[interface{}]interface{} `json: "queueconfig"`
}

func (c *EndpointConfig) Equals(cfg EndpointConfig) bool {
	// @todo build this our more efficiently/explicitely
	return reflect.DeepEqual(*c, cfg)
}
func (c *EndpointConfig) Url() (string, error) {
	host := c.ServiceHost

	if c.ServiceName != "" {
		srvName := fmt.Sprintf("%s.service.consul", c.ServiceName)

		a, err := c.Lb.Next(srvName)
		if err != nil {
			return "", err
		}
		host = fmt.Sprintf("http://%s:%d", a.Address, a.Port)
	}

	return host + c.Uri, nil
}

type RabbitAddress struct {
	KeyRing  []byte
	User     string
	Password string
	Host     string
	Port     uint16
	VHost    string
}

func (a *RabbitAddress) String() string {
	user := a.User
	pass := a.Password

	// if keyring is supplied, decrypt username & password
	if a.KeyRing != nil {
		u, err := secconf.Decode([]byte(a.User), bytes.NewBuffer(a.KeyRing))
		if err != nil {
			user = ""
			log.Printf("Username decryption error: %s", err)
		}
		p, err := secconf.Decode([]byte(a.Password), bytes.NewBuffer(a.KeyRing))
		if err != nil {
			pass = ""
			log.Printf("Password decryption error: %s", err)
		}
		user = string(u[:])
		pass = string(p[:])
	} else {
		log.Println("No keyring supplied, treating rabbitmq credentials as plain text")
	}
	connStr := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", user, pass, a.Host, a.Port, a.VHost)
	return connStr
}

// repo helper
func connectionConfigToAddress(kr []byte, c ConnectionConfig, lb srvlb.SRVLoadBalancerDriver) (RabbitAddress, error) {
	add := RabbitAddress{
		KeyRing:  kr,
		User:     c.User,
		Password: c.Password,
		Host:     c.Host,
		Port:     c.Port,
		VHost:    c.VHost,
	}

	if c.ServiceName != "" {
		a, err := lb.Next("rabbitmq.service.consul")
		if err != nil {
			return add, err
		}
		add.Host = a.Address
		add.Port = a.Port
	}
	log.Printf("rabbitmq address: %s:%d ", add.Host, add.Port)
	return add, nil
}
