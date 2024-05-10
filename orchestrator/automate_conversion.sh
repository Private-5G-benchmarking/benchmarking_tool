#!/bin/bash

# List of folders
folders=("/home/shared/case_study_files/open_source_network/run_1/pcaps" "/home/shared/case_study_files/open_source_network/run_2/pcaps" "/home/shared/case_study_files/open_source_network/run_3/pcaps")

# Iterate over each folder
for index in "${!folders[@]}"; do
    # Increment index by 1 to start from 1
    folder_index=$((index + 1))
   echo "Begun on folder nr ${index}" 

    # Check if the folder exists
    if [ -d "${folders[index]}" ]; then
        # Iterate over each file in the folder
        for file in "${folders[index]}"/*; do
            # Check if the file is a regular file
            if [ -f "$file" ]; then
                # Extract filename without extension
                filename=$(basename -- "$file")
                trial_name="${filename%.*}"
		echo "Begun trial ${trial_name}"
                # Run Ansible playbook with trial_name and run variables
                ansible-playbook /home/shared/benchmarking_tool/orchestrator/playbooks/convert_case_study.yml --extra-vars "trial_name=$trial_name run=$folder_index"
            fi
        done
    else
        echo "Folder ${folders[index]} does not exist."
    fi
done
