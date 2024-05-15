#!/bin/bash

EXTRA_VARS_FILES=("02.yml" "06.yml" "01.yml")

for file in "${EXTRA_VARS_FILES[@]}"; do
    echo "Running ansible playbook with vars $file"
    ansible-playbook playbooks/run_experiment.yml --extra-vars "@/home/shared/benchmarking_tool/orchestrator/trials/$file"

    trial_name=$(sed -n 's/^trial_name: \(.*\)/\1/p' "/home/shared/benchmarking_tool/orchestrator/trials/$file")
    index=5  # Set index to some value
    cp "/opt/hallo/output_files/pcaps/$trial_name.pcapng" "/home/shared/case_study_files/nokia_network/run_$index/pcaps/$trial_name.pcapng"
    ansible-playbook playbooks/convert_case_study.yml --extra-vars "run=$index trial_name=$trial_name"
done

