#!/bin/bash
# Stop on error
set -e

init_environment(){
	if [[ -n $DB_ENV ]]; then
		echo "Environment is : $DB_ENV"
	fi
}

init_environment
