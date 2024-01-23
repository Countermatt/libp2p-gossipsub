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
bootstrap=$8
nbNodes=$((builder + validator))
((nbNodes += regular))
# ========== Prerequisites Install ==========
echo "========== Prerequisites Install =========="
# Install experiment on the grid5000 node for better disk usage
cd /tmp
sudo-g5k
if [ ! -e go1.21.6.linux-amd64.tar.gz ]; then
    # Install Go
    wget "https://go.dev/dl/go1.21.6.linux-amd64.tar.gz"
    sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    # Clone experiment code
    cp -r /home/$login/libp2p-gossipsub /tmp/
    cd /tmp
    cd libp2p-gossipsub
    go build .
fi





# ========== Metrics Gathering Launch ==========
echo "========== Metrics Gathering Launch =========="

# ========== Experiment Launch ==========


echo "========== Experiment Launch =========="

# Run validator
if [ "$validator" -ne 0 ]; then
    for ((i=0; i<$validator; i++)); do
        go run . -duration="$experiment_duration" -nodeType=validator -size="$parcel_size" -bootstrap="$bootstrap" &
        echo "validator $i"
        sleep 0.4
        ((nbNodes -= 1))
   done

    if [ "$builder" -eq 0 ] && [ "$regular" -ne 0 ]; then
        go run . -duration"=$experiment_duration" -nodeType=validator -size="$parcel_size" -bootstrap="$bootstrap" 
    else
        if [ "$validator" -ne 1 ]; then
            go run . -duration="$experiment_duration" -nodeType=validator -size="$parcel_size" -bootstrap="$bootstrap" &
            sleep 0.4
            ((nbNodes -= 1))
        fi
    fi
fi

# Run other nodes
if [ "$regular" -ne 0 ]; then
    for ((i=0; i<$regular; i++)); do
        go run . -duration="$experiment_duration" -nodeType=regular -size="$parcel_size" -bootstrap="$bootstrap" &
        echo "regular $i"
        sleep 0.4
        ((nbNodes -= 1))
    done

    if [ "$builder" -eq 0 ]; then
        go run . -duration="$experiment_duration" -nodeType=regular -size="$parcel_size" -bootstrap="$bootstrap" 
    else
        if [ "$regular" -ne 1 ]; then
            go run . -duration="$experiment_duration" -nodeType=regular -size="$parcel_size" -bootstrap="$bootstrap" &
            sleep 0.4
            ((nbNodes -= 1))
        fi
    fi

fi

if [ "$builder" -ne 0 ]; then
    echo "builder launch"
    go run . -duration="$experiment_duration" -nodeType=builder -size="$parcel_size" -bootstrap="$bootstrap" 
fi

cd /tmp
cd libp2p-gossipsub
cp -r log /home/$login/results/$experiment_name
sleep 30