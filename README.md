# bmstu_team_development

##  Как запуститься  локально?

**создай свой env в корне проекта**

```env
# .env
POSTGRES_USER=admin
POSTGRES_PASSWORD=secret123!
PGADMIN_EMAIL=admin@example.com
PGADMIN_PASSWORD=supersecret!
POSTGRES_DB_NAME=taskdb
POSTGRES_HOST=postgres
```

**запуститься**

```bash 
 docker-compose -f docker-compose.yml -f docker-compose.local.yml up --build
```

**выключить и удалить бд**

```bash 
 docker-compose -f docker-compose.yml -f docker-compose.local.yml down -v
```

**запуск сервиса из командой строки**

```bash
POSTGRES_HOST="localhost" \
POSTGRES_PORT="5432" \
POSTGRES_USER="admin" \
POSTGRES_PASSWORD="secret123!" \
POSTGRES_DB_NAME="taskdb" \
go run ./cmd/main.go
```

**локальный литер**

```bash
golangci-lint run
```

##  Deploy
```bash
./deploy.sh
```

Остановка контейнеров:

```bash 
docker stop $(docker ps -aq)
```

Удаление всех контейнеров:

```bash 
docker rm $(docker ps -aq)
```

Удаление всех образов:

```bash
docker rmi $(docker images -q)
```

Удаление всех неиспользуемых данных:

```bash
docker system prune -a
```

Удаление всех томов:

```bash 
docker volume prune
```
