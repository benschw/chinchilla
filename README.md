# Chinchilla

A service which connects to Rabbitmq queues and delivers messages to REST endpoints.


### terminal 1

	go build
	./chinchilla

### terminal 2

	go run ./example/cmd/server/serve.go


### terminal 3

	go run ./example/cmd/publisher/publish.go -queue demo.foo
	go run ./example/cmd/publisher/publish.go -queue demo.bar
	go run ./example/cmd/publisher/publish.go -queue demo.bar -body "hello galaxy"
