#!/bin/bash
trap "rm server;kill 0" EXIT

go build -o server
#./server -port=8001 &
#./server -port=8002 &
#./server -port=8003 &
./server -addr="localhost:8001" &
./server -addr="localhost:8002" &
./server -addr="localhost:8003" &

sleep 2
echo ">>> start success"

wait
