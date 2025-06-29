#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE DATABASE recordstore;
	CREATE USER recordstoredbo WITH PASSWORD 'rsdbo';
	ALTER DATABASE recordstore OWNER TO recordstoredbo;
EOSQL