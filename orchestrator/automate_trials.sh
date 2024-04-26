#!/bin/bash

EXTRA_VARS_FILES=("01.yml" "02.yml" "03.yml" "04.yml" "07.yml")

for file in "${EXTRA_VARS_FILES[@]}"; do
	echo "Running ansible playbook with vars $file"
	ansible-playbook playbooks/run_experiment.yml --extra-vars "@/home/shared/benchmarking_tool/orchestrator/trials/$file"
done
