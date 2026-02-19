#!/usr/bin/env bash

set -eo pipefail

echo "------------------- START DATABASES --------------------"
docker rm -f dss-crdb-for-migration-testing &> /dev/null || echo "No CRDB to clean up"
echo "Starting CRDB container"
docker run -d --rm --name dss-crdb-for-migration-testing \
	-p 26257:26257 \
  cockroachdb/cockroach:v24.1.3 start-single-node \
  --insecure

docker rm -f dss-ybdb-for-migration-testing &> /dev/null || echo "No YBDB to clean up"
echo "Starting YBDB container"
docker run -d --rm --name dss-ybdb-for-migration-testing \
	-p 5433:5433 \
  yugabytedb/yugabyte:2.25.2.0-b359 bin/yugabyted start --background=false --tserver_flags="ysql_output_buffer_size=1048576"


echo "---------------------- BOOTSTRAP -----------------------"
echo "Building db-manager container for testing"
docker build --rm . -t local-db-manager

count_crdb() {
    docker exec dss-crdb-for-migration-testing \
      ./cockroach sql --insecure -d "${DATABASE}" -e "\d" | wc -l
}

count_ybdb() {
    docker exec dss-ybdb-for-migration-testing \
      sh -c "./bin/ysqlsh -h \$(hostname) \"${DATABASE}\" -c \"\d\" | wc -l"
}

test_datastore() {

    for DATABASE in scd rid aux; do

        if [[ "$DATABASE" == "aux" ]]; then
            DATABASE_FOLDER="aux_"
        else
            DATABASE_FOLDER=$DATABASE
        fi

        echo "--> Working on           ${DATABASE}"

        # shellcheck disable=SC2010
        LATEST_VERSION=$(ls ${MIGRATION_FOLDER}/${DATABASE_FOLDER} | grep -oE 'upto-v[0-9]+\.[0-9]+\.[0-9]+' | sed 's/upto-v//' | sort -V | tail -n 1)

        echo "--> Latest version is    ${LATEST_VERSION}"

        echo "--> Going up"  # Check that we can apply migrations

        docker run --rm --name migration-testing-db-manager \
        --link "${DB_NAME}":db \
        local-db-manager \
        /usr/bin/db-manager migrate \
        --schemas_dir ${SCHEMA_FOLDERS}/"${DATABASE_FOLDER}" \
        --db_version "${LATEST_VERSION}" \
        --datastore_host db \
        --datastore_user ${DB_USER} \
        --datastore_port ${DB_PORT}

        TABLES=$($COUNT_FUNCTION)

        if [ "$TABLES" -eq $EMPTY_COUNT ]; then
            echo "--> Error, database is empty"
            exit 1
        else
            echo "--> Ok, tables in database"
        fi

        echo "--> Going down"  # Check that we can undo migrations

        docker run --rm --name migration-testing-db-manager \
        --link "${DB_NAME}":db \
        local-db-manager \
        /usr/bin/db-manager migrate \
        --schemas_dir ${SCHEMA_FOLDERS}/"${DATABASE_FOLDER}" \
        --db_version "0.0.0" \
        --datastore_host db \
        --datastore_user ${DB_USER} \
        --datastore_port ${DB_PORT}

        # rid table has been renamed in cockroach
        if [[ "$DB_USER" == "root" ]]; then
        if [[ "$DATABASE" == "rid" ]]; then
            DATABASE="defaultdb"
        fi
        fi

        TABLES=$($COUNT_FUNCTION)

        if [ "$TABLES" -eq $EMPTY_COUNT ]; then
            echo "--> Ok, database is empty"
        else
            echo "--> Err, leftover tables"
            echo "$TABLES"
            exit 1
        fi

        echo "--> Going up again"  # Check that we can apply migrations again

        docker run --rm --name migration-testing-db-manager \
        --link "${DB_NAME}":db \
        local-db-manager \
        /usr/bin/db-manager migrate \
        --schemas_dir ${SCHEMA_FOLDERS}/"${DATABASE_FOLDER}" \
        --db_version "${LATEST_VERSION}" \
        --datastore_host db \
        --datastore_user ${DB_USER} \
        --datastore_port ${DB_PORT}
    done
}

echo "---------------------- COCKROACH -----------------------"

DB_NAME=dss-crdb-for-migration-testing
DB_PORT=26257
DB_USER=root
MIGRATION_FOLDER=build/db_schemas
SCHEMA_FOLDERS=db-schemas
COUNT_FUNCTION=count_crdb
EMPTY_COUNT=2

test_datastore

echo "---------------------- YUGABYTEDB ----------------------"

DB_NAME=dss-ybdb-for-migration-testing
DB_PORT=5433
DB_USER=yugabyte
MIGRATION_FOLDER=build/db_schemas/yugabyte
SCHEMA_FOLDERS=db-schemas/yugabyte
COUNT_FUNCTION=count_ybdb
EMPTY_COUNT=0

test_datastore
