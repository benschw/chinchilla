package ep

import "reflect"

type Config struct {
	Endpoints []EndpointConfig `json: "endpoints"`
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
