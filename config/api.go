package config

import (
	"fmt"
	"log"
	"reflect"

	"github.com/benschw/dns-clb-go/clb"
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
}

type EndpointConfig struct {
	Lb          clb.LoadBalancer
	Name        string `json: "name"`
	ServiceHost string `json: "servicehost"`
	ServiceName string `json: "servicename"`
	Uri         string `json: "uri"`
	Method      string `json: "method"`
	QueueName   string `json: "queuename"`
	Prefetch    int    `json: "prefetch"`
}

func (c *EndpointConfig) Equals(cfg EndpointConfig) bool {
	// @todo build this our more efficiently/explicitely
	return reflect.DeepEqual(*c, cfg)
}
func (c *EndpointConfig) Url() (string, error) {
	host := c.ServiceHost

	if c.ServiceName != "" {
		srvName := fmt.Sprintf("%s.service.consul", c.ServiceName)

		a, err := c.Lb.GetAddress(srvName)
		if err != nil {
			return "", err
		}
		host = fmt.Sprintf("http://%s:%d", a.Address, a.Port)
	}

	return host + c.Uri, nil
}

type RabbitAddress struct {
	User     string
	Password string
	Host     string
	Port     uint16
}

func (a *RabbitAddress) String() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/", a.User, a.Password, a.Host, a.Port)
}

// repo helper
func connectionConfigToAddress(c ConnectionConfig, lb clb.LoadBalancer) (RabbitAddress, error) {
	add := RabbitAddress{
		User:     c.User,
		Password: c.Password,
		Host:     c.Host,
		Port:     c.Port,
	}

	if c.ServiceName != "" {
		a, err := lb.GetAddress("rabbitmq.service.consul")
		if err != nil {
			return add, err
		}
		add.Host = a.Address
		add.Port = a.Port
	}
	log.Printf("Using %s:%d to connect to rabbitmq", add.Host, add.Port)
	return add, nil
}
