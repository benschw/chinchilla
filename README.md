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




## testing

### Install anc configure Rabbitmq

	sudo aptitude

	# allow guest login other than loopback (for consul discovery)
	cat << EOF > /etc/rabbitmq/rabbitmq.config
	[{rabbit,[
		{loopback_users, []}
	]}].
	EOF

### Install and configure Consul

	# get the consul binary
	wget https://dl.bintray.com/mitchellh/consul/0.5.2_linux_amd64.zip
	unzip 0.5.2_linux_amd64.zip

	# get the web ui and drop it in your /tmp dir
	wget https://dl.bintray.com/mitchellh/consul/0.5.2_web_ui.zip
	unzip 0.5.2_web_ui.zip
	mv dist /tmp/web-ui


	# run wrapper script that will configure some demo info
	./consul.sh
