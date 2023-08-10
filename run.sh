#!/bin/bash

# Bash Script to launch Gossipsub PANDAS

# ========== Parameters ==========
experiment_duration=$1
experiment_name=$2
builder=$3
validator=$4
regular=$5
login=$6
metrics_file="$(hostname)-log"
parcel_size=$7
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
sleep 1
sar -A -o $metrics_file 1 $exp_duration >/dev/null 2>&1 &
sleep 1

# ========== Experiment Launch ==========
# Run builder


# Run validator
if [ "$validator" -ne 0 ]; then
    for ((i=0; i<$validator-1; i++)); do
        go run . -duration="$experiment_duration" -nodeType=validator -size="$parcel_size" &
    done

    if [ "$builder" -eq 0 ] && [ "$regular" -ne 0 ]; then
        go run . -duration="$experiment_duration" -nodeType=validator -size="$parcel_size"
    else
        go run . -duration="$experiment_duration" -nodeType=validator -size="$parcel_size" &
    fi
fi

# Run other nodes
if [ "$regular" -ne 0 ]; then
    for ((i=0; i<$regular-1; i++)); do
        go run . -duration="$experiment_duration" -nodeType=nonvalidator -size="$parcel_size" &
    done

    if [ "$builder" -eq 0 ]; then
        go run . -duration="$experiment_duration" -nodeType=nonvalidator -size="$parcel_size"
    else
        go run . -duration="$experiment_duration" -nodeType=nonvalidator -size="$parcel_size" &
    fi

fi

if [ "$builder" -ne 0 ]; then
    go run . -duration="$experiment_duration" -nodeType=builder -size="$parcel_size"
fi

cp *.csv /home/$login/results/$experiment_name/
sleep 1
cp $metrics_file /home/$login/results/$experiment_name/
sleep 1

rm go1.20.4.linux-amd64.tar.gz
rm -rf *.csv
rm -rf *-log