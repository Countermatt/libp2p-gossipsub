#!/bin/bash


#Install go
cd /tmp
wget "https://go.dev/dl/go1.20.4.linux-amd64.tar.gz"
sudo-g5k tar -C /usr/local -xzf go1.20.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

cd /tmp/libp2p-das

go run . -duration $1 


mkdir "$2$(date +%d-%m-%y-%H-%M)"
cp *.csv "$2$(date +%d-%m-%y-%H-%M)"