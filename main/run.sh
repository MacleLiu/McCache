#!/bin/bash
trap "rm server;kill 0" EXIT

go build -o server

./server -addr="localhost:8001" -etcdAddr="127.0.0.1:2379" &
./server -addr="localhost:8002" -etcdAddr="127.0.0.1:2379" &
./server -addr="localhost:8003" -etcdAddr="127.0.0.1:2379" &

sleep 2
echo ">>> start success"

wait
