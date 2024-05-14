#!/bin/bash

# Remove the space after the equal sign in the variable assignment
NUMBER_OF_RUNS=(1 2 3 4 5)

for index in "${NUMBER_OF_RUNS[@]}"; do
    echo "index is $index"

    EXTRA_VARS_FILES=("02.yml" "06.yml" "01.yml") 

    for file in "${EXTRA_VARS_FILES[@]}"; do
        echo "Running ansible playbook with vars $file"
        ansible-playbook playbooks/run_experiment.yml --extra-vars "@/home/shared/benchmarking_tool/orchestrator/trials/$file"
        
        trial_name=$(sed -n 's/^trial_name: \(.*\)/\1/p' "/home/shared/benchmarking_tool/orchestrator/trials/$file")
        ansible-playbook playbooks/convert_case_study.yml --extra-vars "run=$index trial_name=$trial_name"
    done
done
