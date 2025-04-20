#!/bin/bash

set -e  # Остановить скрипт при любой ошибке

echo "🚀 Pull latest image from Docker Hub..."
docker pull volnamax1/todolist:latest

echo "♻️ Recreate containers using docker-compose.prod.yml..."
docker-compose -f docker-compose.prod.yml up -d

echo "✅ Deployment complete! App should be available on port 8080"
