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
vhost: /
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

read -r -d '' TOPIC_CFG << EOF
name: Topic
servicename: foo
uri: /foo
method: POST
consumerstrategy: topic
queueconfig:
  prefetch: 5
  topicname: foo.*
  queuename: all-foos
  exchangename: demo
EOF

read -r -d '' TOPIC2_CFG << EOF
name: Topic2
servicename: foo
uri: /foo
method: POST
consumerstrategy: topic
queueconfig:
  prefetch: 5
  topicname: foo.*
  queuename: more-foos
  exchangename: demo
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

# Repeater Config
read -r -d '' DC1_CFG << EOF
name: dc1
user: guest
password: guest
vhost: /
host: localhost
port: 5672
EOF
#servicename: rabbitmq

read -r -d '' REP_CFG << EOF
name: Repeater
servicename: foo
uri: /foo
method: POST
consumerstrategy: topic
deliverystrategy: topic-repeater
queueconfig:
  prefetch: 5
  topicname: '#'
  queuename: repeater-queue
  exchangename: demo-in
  connection: dc1
  exchangeout: demo-out
EOF

read -r -d '' REP_OUT_CFG << EOF
name: Repeated
servicename: foo
uri: /foo
method: POST
consumerstrategy: topic
queueconfig:
  prefetch: 5
  topicname: '#'
  queuename: repeated
  exchangename: demo-out
EOF

curl -X PUT http://127.0.0.1:8500/v1/kv/chinchilla/endpoints/repeater.yaml -d "$REP_CFG"
curl -X PUT http://127.0.0.1:8500/v1/kv/chinchilla/repeater/connections/dc1.yaml -d "$DC1_CFG"
curl -X PUT http://127.0.0.1:8500/v1/kv/chinchilla/endpoints/repeated.yaml -d "$REP_OUT_CFG"

# End

echo "Configuring"
curl -X PUT http://127.0.0.1:8500/v1/agent/service/register -d "$RABBIT_SVC"
curl -X PUT http://127.0.0.1:8500/v1/agent/service/register -d "$FOO_SVC"

curl -X PUT http://127.0.0.1:8500/v1/kv/chinchilla/connection.yaml -d "$CONN_CFG"
curl -X PUT http://127.0.0.1:8500/v1/kv/chinchilla/endpoints/foo.yaml -d "$FOO_CFG"
curl -X PUT http://127.0.0.1:8500/v1/kv/chinchilla/endpoints/topic.yaml -d "$TOPIC_CFG"
curl -X PUT http://127.0.0.1:8500/v1/kv/chinchilla/endpoints/topic2.yaml -d "$TOPIC2_CFG"





wait
