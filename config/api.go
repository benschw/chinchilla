package config

import (
	"fmt"
	"reflect"
)

type Config struct {
	Connection ConnectionConfig `json: "connection"`
	Endpoints  []EndpointConfig `json: "endpoints"`
}

type ConnectionConfig struct {
	User        string `json: "user"`
	Password    string `json: "password"`
	Host        string `json: "host"`
	ServiceName string `json: "servicename"`
	Port        int    `json: "port"`
}

type EndpointConfig struct {
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

type RabbitAddress struct {
	User     string
	Password string
	Host     string
	Port     int
}

func (a *RabbitAddress) String() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/", a.User, a.Password, a.Host, a.Port)
}

// repo helper
func connectionConfigToAddress(c ConnectionConfig) (RabbitAddress, error) {
	// @todo discover if ServiceName is set
	return RabbitAddress{
		User:     c.User,
		Password: c.Password,
		Host:     c.Host,
		Port:     c.Port,
	}, nil
}
