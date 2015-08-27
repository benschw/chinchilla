package ep

import "reflect"

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
}

func (c *EndpointConfig) Equals(cfg EndpointConfig) bool {
	// @todo build this our more efficiently/explicitely
	return reflect.DeepEqual(*c, cfg)
}
