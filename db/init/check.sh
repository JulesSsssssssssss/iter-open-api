#/bin/bash
# Stop on error
set -e

init_environment(){
	if [[ -n $DB_ENV ]]; then
		echo "Environment is : $DB_ENV\n"
	fi
	# Create tables if not exist
	if [[ $DB_ENV == "dev" ]]; then
		# seed the db
	fi
}

check_environment()
