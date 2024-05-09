#!/bin/bash

pps_rates=(10000 20000 30000 40000 50000 60000 70000 80000 90000 100000)
payload_sizes=(16 426 1432)
for payload_size in "${payload_sizes[@]}"; do
    echo "Running trials for payload size: $payload_size"
    # Iterate over the values and run Ansible playbook
    for pps in "${pps_rates[@]}"; do
        echo "Running Ansible playbook for value: $pps"
        ansible-playbook validate_bridge.yml --extra-vars "pps=$pps payload_size=$payload_size"
    done
done
