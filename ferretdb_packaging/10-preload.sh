#!/bin/bash

set -e

# https://github.com/microsoft/documentdb/tree/main/pg_documentdb/src/configs

cat <<EOT >> $PGDATA/postgresql.conf
shared_preload_libraries                  = 'pg_cron,pg_documentdb_core,pg_documentdb'
cron.database_name                        = 'postgres'

documentdb.enableSchemaValidation         = true
documentdb.enableBypassDocumentValidation = true

documentdb.enableUserCrud                 = true
documentdb.maxUserLimit                   = 100
EOT

source /usr/local/bin/docker-entrypoint.sh

# restarting is needed to create extension
docker_temp_server_stop
docker_temp_server_start "$@"
