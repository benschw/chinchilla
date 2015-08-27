#!/bin/bash

# wget https://dl.bintray.com/mitchellh/consul/0.5.2_linux_amd64.zip
# unzip 0.5.2_linux_amd64.zip
# wget https://dl.bintray.com/mitchellh/consul/0.5.2_web_ui.zip
# unzip 0.5.2_web_ui.zip
# mv dist /tmp/web-ui

./consul agent -server -bootstrap-expect 1 -data-dir /tmp/consul -ui-dir /tmp/web-ui &

sleep 5

echo "Configuring \"Foo\""
curl -X PUT http://localhost:8500/v1/kv/chinchilla/
curl -X PUT http://localhost:8500/v1/kv/chinchilla/endpoints/
curl -X PUT http://localhost:8500/v1/kv/chinchilla/endpoints/Foo/
curl -X PUT http://localhost:8500/v1/kv/chinchilla/endpoints/Foo/Name -d "Foo"
curl -X PUT http://localhost:8500/v1/kv/chinchilla/endpoints/Foo/ServiceHost -d "http://localhost:8080"
curl -X PUT http://localhost:8500/v1/kv/chinchilla/endpoints/Foo/Uri -d "/foo"
curl -X PUT http://localhost:8500/v1/kv/chinchilla/endpoints/Foo/Method -d "POST"
curl -X PUT http://localhost:8500/v1/kv/chinchilla/endpoints/Foo/QueueName -d "demo.foo"
curl -X PUT http://localhost:8500/v1/kv/chinchilla/endpoints/Foo/Enable -d "true"

wait
