#!/bin/bash

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
API_KEY=$API_KEY

echo "Connecting to $DB_HOST:$DB_PORT"
echo "Using API key: $API_KEY"
