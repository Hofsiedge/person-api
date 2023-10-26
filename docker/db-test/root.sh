# exit on error immediately
set -e

{
	. /usr/local/bin/docker-entrypoint.sh
	docker_verify_minimum_env
	ls /docker-entrypoint-initdb.d/
	docker_init_database_dir
	export PGPASSWORD="$POSTGRES_PASSWORD"
	docker_temp_server_start
	docker_setup_db
	docker_process_init_files /docker-entrypoint-initdb.d/*

	psql --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
		-c "CREATE EXTENSION pgtap; CREATE SCHEMA test;"
} > /dev/null

echo "PosgtgreSQL server is initialized, applying migrations..."

# apply migrations
migrate \
	-path "/migrations/" \
	-database "postgres:///$POSTGRES_DB?host=/var/run/postgresql/&user=$POSTGRES_USER&password=$POSTGRES_PASSWORD" \
	up

# which pg_prove

# run tests
pg_prove \
	-U "$POSTGRES_USER" \
	-d "$POSTGRES_DB" \
	--schema test \
	--runtests \
	--failures \
	--verbose

docker_temp_server_stop > /dev/null
