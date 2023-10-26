# exit on error immediately
set -e

{
	. /usr/local/bin/docker-entrypoint.sh
	docker_setup_env
	docker_create_db_directories
} > /dev/null

su -s /bin/bash -c ". /root.sh" postgres 
