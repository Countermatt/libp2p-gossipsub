#!/bin/bash

# Bash Script to launch Gossipsub PANDAS

# ========== Parameters ==========
experiment_duration=$1
experiment_name=$2
builder=$3
validator=$4
regular=$5
login=$6
experiment_folder="/home/$login/$experiment_name"
metrics_file= "$(hostname)-log"
# ========== Prerequisites Install ==========

# Install experiment on the grid5000 node for better disk usage
cd /tmp

# Install Go
wget "https://go.dev/dl/go1.20.4.linux-amd64.tar.gz"
sudo-g5k tar -C /usr/local -xzf go1.20.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Clone experiment code
git clone https://github.com/Countermatt/libp2p-gossipsub.git
cd libp2p-gossipsub
go build 

# ========== Metrics Gathering Launch ==========

sudo-g5k systemctl start sysstat
mkdir "$experiment_folder"
sar -A 1 $experiment_duration > $(hostname)-log &

# ========== Experiment Launch ==========
# Run builder
if [ "$builder" -ne 0 ]; then
    for ((i=0; i<$builder-1; i++)); do
        go run . -duration="$experiment_duration" -nodeType=builder &
    done

    if [ "$validator" == "0" ] && [ "$regular" == "0" ]; then
        go run . -duration="$experiment_duration" -nodeType=builder
    else
        go run . -duration="$experiment_duration" -nodeType=builder &
    fi
fi

# Run validator
if [ "$validator" -ne "0" ]; then
    for ((i=0; i<$validator-1; i++)); do
        go run . -duration="$experiment_duration" -nodeType=validator &
        sleep 0.1
    done

    if [ "$regular" == "0" ]; then
        go run . -duration="$experiment_duration" -nodeType=validator
    else
        go run . -duration="$experiment_duration" -nodeType=validator &
    fi
fi

# Run other nodes
if [ "$regular" -ne 0 ]; then
    for ((i=0; i<$regular-1; i++)); do
        go run . -duration="$experiment_duration" -nodeType=nonvalidator &
        sleep 0.1
    done
    go run . -duration="$experiment_duration" -nodeType=nonvalidator
fi

mkdir "$experiment_folder"
cp *.csv $experiment_folder
cp *.txt $experiment_folder