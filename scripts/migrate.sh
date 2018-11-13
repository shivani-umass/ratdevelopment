#!/bin/bash
# Migrate the schema to the Cassandra DB
# Resolve the directory of this script

SCRIPT_DIR=$(dirname $(readlink -f "$0"))
until cqlsh "$1" -e 'desc schema'; do
    sleep 5
done
cqlsh "$1" --file "$SCRIPT_DIR/schema.cql"
