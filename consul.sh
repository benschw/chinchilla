#!/bin/bash

# wget https://dl.bintray.com/mitchellh/consul/0.5.2_linux_amd64.zip
# unzip 0.5.2_linux_amd64.zip
# wget https://dl.bintray.com/mitchellh/consul/0.5.2_web_ui.zip
# unzip 0.5.2_web_ui.zip
# mv dist /tmp/web-ui

./consul agent -server -bootstrap-expect 1 -data-dir /tmp/consul -ui-dir /tmp/web-ui &

sleep 5

USER=$(./chinchilla -keyring ./test-keys/.pubring.gpg encrypt guest)
PASS=$(./chinchilla -keyring ./test-keys/.pubring.gpg encrypt guest)

read -r -d '' CONN_CFG << EOF
user: $USER
password: $PASS
servicename: rabbitmq
EOF

read -r -d '' FOO_CFG << EOF
name: Foo
servicename: foo
uri: /foo
method: POST
queueconfig:
  prefetch: 5
  queuename: demo.foo
EOF

read -r -d '' RABBIT_SVC << EOF
{
  "ID": "rabbitmq1",
  "Name": "rabbitmq",
  "Address": "127.0.0.1",
  "Port": 5672
}
EOF
read -r -d '' FOO_SVC << EOF
{
  "ID": "foo1",
  "Name": "foo",
  "Address": "127.0.0.1",
  "Port": 8080
}
EOF

echo "Configuring"
curl -X PUT http://localhost:8500/v1/agent/service/register -d "$RABBIT_SVC"
curl -X PUT http://localhost:8500/v1/agent/service/register -d "$FOO_SVC"

curl -X PUT http://localhost:8500/v1/kv/chinchilla/connection.yaml -d "$CONN_CFG"
curl -X PUT http://localhost:8500/v1/kv/chinchilla/endpoints/foo.yaml -d "$FOO_CFG"






wait
