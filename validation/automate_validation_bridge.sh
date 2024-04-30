#!/bin/bash

pps_rates=(1000 10000 20000 30000 40000 50000 60000 70000 80000 90000 100000 110000 120000 130000 140000 150000 160000 170000 180000 190000 200000)

# Iterate over the values and run Ansible playbook
for pps in "${pps_rates[@]}"; do
    echo "Running Ansible playbook for value: $value"
    ansible-playbook validate_bridge.yml --extra-vars "pps=$pps"
done
