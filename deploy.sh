#!/bin/bash

set -e  # Остановить скрипт при любой ошибке

DOCKER_USER="volnamax1"
COMPOSE_PROD="docker-compose -f docker-compose.yml -f docker-compose.prod.yml"

echo "Pulling the latest image from Docker Hub..."
docker pull $DOCKER_USER/todolist:latest

echo "Recreating containers using docker-compose.prod.yml..."
$COMPOSE_PROD down -v
$COMPOSE_PROD up -d

echo "Deployment complete! App should be available on port 8080"
