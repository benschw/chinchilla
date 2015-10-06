[![Build Status](https://travis-ci.org/benschw/chinchilla.svg?branch=master)](https://travis-ci.org/benschw/chinchilla)
[Download](https://drone.io/github.com/benschw/chinchilla/files/chinchilla.gz)

# Chinchilla

A service which connects to Rabbitmq queues and delivers messages to REST endpoints.

Chinchilla can be configured either with a yaml config, or with consul. In either case, the
config backend is watched for changes and will live-update your running chinchilla daemon to
reflect the new configuration.

## local demo

	wget https://dl.bintray.com/mitchellh/consul/0.5.2_linux_amd64.zip
	unzip 0.5.2_linux_amd64.zip
	wget https://dl.bintray.com/mitchellh/consul/0.5.2_web_ui.zip
	unzip 0.5.2_web_ui.zip
	mv dist /tmp/web-ui

### terminal 1

	go build
	./consul.sh

### terminal 2

	 SRVLB_HOST=127.0.0.1:8600 ./chinchilla


### terminal 3

	go run ./example/cmd/server/serve.go


### terminal 4

	go run ./example/cmd/publisher/publish.go -queue demo.foo
	go run ./example/cmd/publisher/publish.go -queue demo.bar
	go run ./example/cmd/publisher/publish.go -queue demo.bar -body "hello galaxy"




## testing

### Install and configure Rabbitmq

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

### Encrypt Rabbit Credentials in your Config
see [https://github.com/xordataexchange/crypt] for details


Update `app.batch` with your info, and run the following to generate your keys:

	gpg2 --batch --armor --gen-key app.batch

This will generate `.pubring.gpg` and `.secring.gpg` for encrypting and
decrypting rabbitmq credentials in your configuration backend.

#### app.batch

	%echo Generating a configuration OpenPGP key
	Key-Type: default
	Subkey-Type: default
	Name-Real: app
	Name-Comment: app configuration key
	Name-Email: app@example.com
	Expire-Date: 0
	%pubring .pubring.gpg
	%secring .secring.gpg
	%commit
	%echo done
