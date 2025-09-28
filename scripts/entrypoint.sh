#!/usr/bin/env bash
set -e

echo "Waiting for Postgres..."
./wait-for-it.sh postgres:5432 -- echo "Postgres is ready"

echo "Waiting for Redis..."
./wait-for-it.sh redis:6379 -- echo "Redis is ready"

echo "Starting Go service..."
exec /app/main
