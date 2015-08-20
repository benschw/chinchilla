# Chinchilla

A service which connects to Rabbitmq queues and delivers messages to REST endpoints.


### terminal 1

	go build
	./chinchilla


### terminal 2

	go run ./example/send.go -queue demo.foo
	go run ./example/send.go -queue demo.bar
