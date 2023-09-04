#!/bin/bash

source "venv/bin/activate"
cp libp2p-gossipsub/run.sh .
python3 libp2p-gossipsub/experiment_launch.py
sleep 7200