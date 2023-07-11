#!/bin/bash


#Install go
cd /tmp
wget "https://go.dev/dl/go1.20.4.linux-amd64.tar.gz"
sudo-g5k tar -C /usr/local -xzf go1.20.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

git clone https://github.com/Countermatt/libp2p-gossipsub.git

cd libp2p-gossipsub

go run . -duration $1 

if [!-d /home/mapigaglio/result]; then
    mkdir -p /home/mapigaglio/result;
fi;

mkdir "$2$(date +%d-%m-%y-%H-%M)"
cp *.csv "$2$(date +%d-%m-%y-%H-%M)"

