package ep

type Config struct {
	Endpoints []EndpointConfig `json: "endpoints"`
}

type EndpointConfig struct {
	Name        string `json: "name"`
	ServiceHost string `json: "servicehost"`
	ServiceName string `json: "servicename"`
	Uri         string `json: "uri"`
	QueueName   string `json: "queuename"`
}
