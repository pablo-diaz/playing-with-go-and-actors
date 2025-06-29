#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username recordstoredbo --dbname recordstore <<-EOSQL
    CREATE TABLE albums (
        id varchar(100) not null PRIMARY KEY,
        title varchar(200) not null,
        artist varchar(200) not null,
        price varchar(50) not null
    );
EOSQL

