#!/bin/bash
trap "rm apiserver;kill 0" EXIT

go build -o apiserver
./apiserver -pattern="/mccache" -addr="localhost:1234" -etcdAddr="127.0.0.1:2379" &
#./apiserver

sleep 2
echo ">>> apiserver start success"

wait
