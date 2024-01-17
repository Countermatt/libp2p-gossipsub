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
echo "========== Prerequisites Install =========="
# Install experiment on the grid5000 node for better disk usage
#cd /tmp

# Install Go
#wget "https://go.dev/dl/go1.21.6.linux-amd64.tar.gz"
#sudo-g5k tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
#export PATH=$PATH:/usr/local/go/bin

# Clone experiment code
#git clone https://github.com/Countermatt/libp2p-gossipsub.git
cp -r ../libp2p-gossipsub /tmp/
cd /tmp
cd libp2p-gossipsub
go build

# ========== Metrics Gathering Launch ==========
echo "========== Metrics Gathering Launch =========="
#sudo-g5k systemctl start sysstat
sleep 1
sar -A -o $metrics_file 1 $experiment_duration >/dev/null 2>&1 &
sleep 1

# ========== Experiment Launch ==========

echo "========== Experiment Launch =========="

# Run validator
if [ "$validator" -ne 0 ]; then
    for ((i=0; i<$validator; i++)); do
        go run . -duration="$experiment_duration" -nodeType=validator -size="$parcel_size" &
        echo "validator $i"
        sleep 0.1
   done

    if [ "$builder" -eq 0 ] && [ "$regular" -ne 0 ]; then
        go run . -duration"$experiment_duration" -nodeType=validator -size="$parcel_size"
    else
        if [ "$validator" -ne 1 ]; then
            go run . -duration="$experiment_duration" -nodeType=validator -size="$parcel_size" &
            sleep 0.1
        fi
    fi
fi

# Run other nodes
if [ "$regular" -ne 0 ]; then
    for ((i=0; i<$regular; i++)); do
        go run . -duration="$experiment_duration" -nodeType=nonvalidator -size="$parcel_size" &
        echo "regular $i"
        sleep 0.1
    done

    if [ "$builder" -eq 0 ]; then
        go run . -duration="$experiment_duration" -nodeType=nonvalidator -size="$parcel_size"
    else
        if [ "$regular" -ne 1 ]; then
            go run . -duration="$experiment_duration" -nodeType=nonvalidator -size="$parcel_size" &
            sleep 0.1
        fi
    fi

fi

if [ "$builder" -ne 0 ]; then
    echo "builder launch"
    go run . -duration="$experiment_duration" -nodeType=builder -size="$parcel_size"
fi


# echo "========== Log copy =========="

# directory=$(pwd)
# target_count=$((($builder + $validator + $regular) * 2))  # Change this to the desired number of files

# while true; do
#     file_count=$(find "$directory" -type f -name "*.log" | wc -l)
#     if [ "$file_count" -ge "$target_count" ]; then
#         echo "Found $file_count files. Exiting loop."
#         break
#     else
#         echo "Found $file_count files. Waiting for $target_count files..."
#         sleep 5  # Adjust the sleep interval as needed
#     fi
# done
# cp *.log /home/$login/results/$experiment_name/
# sleep 1
# cp $metrics_file /home/$login/results/$experiment_name/
# sleep 1
# cd /tmp
# #rm -rf libp2p-gossipsub
# #rm go1.20.4.linux-amd64.tar.gz
