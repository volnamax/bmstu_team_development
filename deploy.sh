#!/bin/bash

set -e  # Остановить скрипт при любой ошибке

echo "🚀 Pulling the latest image from Docker Hub..."
docker pull volnamax1/todolist:latest

echo "♻️ Recreating containers using docker-compose.prod.yml..."
docker-compose -f docker-compose.prod.yml down  # Останавливаем и удаляем старые контейнеры
docker-compose -f docker-compose.prod.yml up -d  # Перезапускаем контейнеры

echo "✅ Deployment complete! App should be available on port 8080"
