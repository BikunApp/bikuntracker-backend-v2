#!/bin/sh

echo "Running database migrations..."
./migrate

echo "Seeding database with bus data..."
./seeder bus

echo "Starting backend application..."
exec ./exec
