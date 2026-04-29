# Effective Mobile Test Project

REST-сервис для агрегации данных об онлайн-подписках пользователей.

Исходный текст тестового задания вынесен в [TEST_TASK.md](./TEST_TASK.md).

## Что реализовано

- CRUDL API для подписок
- подсчет суммарной стоимости подписок за период с фильтрацией по `user_id` и `service_name`
- PostgreSQL в качестве хранилища
- миграции для инициализации базы
- логирование запросов и ошибок
- конфигурация через `.env`
- Swagger-документация
- запуск через `docker compose`

## Стек

- Go
- Echo
- PostgreSQL
- pgx
- golang-migrate
- Swagger (`swaggo`)
- Docker Compose

## Структура

- `cmd/app` - основной HTTP-сервис
- `cmd/migrator` - приложение для запуска миграций
- `internal/http/handler` - HTTP-обработчики и DTO
- `internal/service` - бизнес-логика
- `internal/repository/postgres` - работа с PostgreSQL
- `deployments/migrations/postgreSQL` - SQL-миграции
- `deployments/envs/prod.env` - основной env-конфиг для запуска через `Makefile`
- `deployments/envs/dev.env` - альтернативный env-конфиг для локальной разработки

## Запуск

Для локального запуска через Docker Compose:

```bash
make build-up
```

Или напрямую:

```bash
docker compose -f docker-compose.yml --env-file ./deployments/envs/prod.env up -d --build
```

Остановка контейнеров:

```bash
make down
```

Сервис по умолчанию будет доступен на:

```text
http://localhost:8080
```

## Swagger

Swagger UI доступен по адресу:

```text
http://localhost:8080/swagger/index.html
```

Для регенерации swagger-документации:

```bash
make swagger
```

## Основные ручки

- `POST /subscriptions` - создать подписку
- `GET /subscriptions?limit=100&offset=0` - получить список подписок с пагинацией
- `GET /subscriptions/{id}` - получить подписку по ID
- `PUT /subscriptions/{id}` - обновить подписку по ID
- `PATCH /subscriptions/{id}` - частично обновить подписку по ID
- `DELETE /subscriptions/{id}` - удалить подписку по ID
- `GET /subscriptions/total` - посчитать суммарную стоимость подписок за период
- `GET /health` - healthcheck сервиса

Пример запроса на создание подписки:

```json
{
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025"
}
```

Пример запроса списка подписок:

```text
GET /subscriptions?limit=100&offset=0
```

Пример частичного обновления подписки:

```json
{
  "price": 500,
  "end_date": "12-2025"
}
```

Для очистки даты окончания:

```json
{
  "clear_end_date": true
}
```

Пример запроса на подсчет суммы:

```text
GET /subscriptions/total?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus&from=07-2025&to=09-2025
```

## Конфигурация

Для запуска используется env-файл:

```text
./deployments/envs/prod.env
```

Основные переменные:

- `HOST`
- `PORT`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `POSTGRES_DB`
- `POSTGRES_PORT`
- `DB_CONNECT_TIMEOUT`
- `REQUEST_TIMEOUT`
- `DB_QUERY_TIMEOUT`

## Миграции

SQL-миграции лежат в:

```text
deployments/migrations/postgreSQL
```

При запуске через `docker compose` поднимается отдельный контейнер `migrator`, который применяет миграции к PostgreSQL.

## Тесты

Запуск тестов:

```bash
go test ./...
```

Через `Makefile`:

```bash
make test
```

Интеграционные тесты:

```bash
make test-integration
```

Интеграционные тесты используют Docker/Testcontainers для проверки PostgreSQL-слоя.
