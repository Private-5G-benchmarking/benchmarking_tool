#!/bin/bash

# Define the list of values
values=(1000 10000 20000 30000 40000 50000 60000 70000 80000 90000 100000)

# Iterate over the values and run Ansible playbook
for value in "${values[@]}"; do
    echo "Running Ansible playbook for value: $value"
    ansible-playbook validate_bridge.yml --extra-vars "pps=$value"
done
