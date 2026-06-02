# TaskBoard API

## Возможности

* создание и получение пользователей;
* создание и получение проектов;
* добавление участников в проект;
* закрытие проекта;
* создание задач;
* получение задач;
* изменение статуса задач;
* удаление задач;
* добавление и получение комментариев к задачам.

## Архитектура

```text
HTTP Handler
    ↓
Service
    ↓
Repository
    ↓
PostgreSQL
```

## Технологии

* Go
* Chi Router
* PostgreSQL
* pgxpool
* Squirrel
* Goose Migrations
* Docker Compose
* OpenAPI
* oapi-codegen

### Сборка проекта

```powershell
go build ./...
```

---

### Запуск unit-тестов

```powershell
go test ./internal/service
```

### Запуск интеграционного теста

```powershell
go test -tags=integration ./internal/app
```

### Запуск всех тестов

```powershell
go test ./...
```

## Запуск базы данных

```bash
docker compose up -d
```

## Применение миграций

```bash
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5432/taskboard?sslmode=disable" up
```

## Запуск приложения

```bash
go run ./cmd/server
```

После запуска сервер доступен по адресу:

```text
http://localhost:8080
```

Проверка:

```bash
curl http://localhost:8080/health
```

## Основные endpoint'ы

```text
POST   /api/v1/users
GET    /api/v1/users/{userID}

POST   /api/v1/projects
GET    /api/v1/projects
GET    /api/v1/projects/{projectID}
POST   /api/v1/projects/{projectID}/members
PATCH  /api/v1/projects/{projectID}/close

POST   /api/v1/tasks
GET    /api/v1/tasks
GET    /api/v1/tasks/{taskID}
PATCH  /api/v1/tasks/{taskID}/status
DELETE /api/v1/tasks/{taskID}

POST   /api/v1/tasks/{taskID}/comments
GET    /api/v1/tasks/{taskID}/comments
```
# Примеры запросов (PowerShell)

## Проверка сервера

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/health" `
-Method GET
```

---

## Создание пользователя

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/api/v1/users" `
-Method POST `
-ContentType "application/json" `
-Body '{"name":"Miroslav","email":"miroslav@test.com"}'
```

---

## Получение пользователя

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/api/v1/users/1" `
-Method GET
```

---

## Создание проекта

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/api/v1/projects" `
-Method POST `
-ContentType "application/json" `
-Body '{"name":"TaskBoard","description":"Go course project"}'
```

---

## Получение проекта

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/api/v1/projects/1" `
-Method GET
```

---

## Получение списка проектов

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/api/v1/projects" `
-Method GET
```

---

## Добавление участника в проект

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/api/v1/projects/1/members" `
-Method POST `
-ContentType "application/json" `
-Body '{"user_id":1,"role":"member"}'
```

---

## Создание задачи

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/api/v1/tasks" `
-Method POST `
-ContentType "application/json" `
-Body '{"project_id":1,"assignee_id":1,"title":"First task","description":"Test task"}'
```

---

## Получение задачи

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/api/v1/tasks/1" `
-Method GET
```

---

## Получение списка задач

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/api/v1/tasks" `
-Method GET
```

---

## Перевод задачи в статус IN_PROGRESS

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/api/v1/tasks/1/status" `
-Method PATCH `
-ContentType "application/json" `
-Body '{"status":"in_progress"}'
```

---

## Перевод задачи в статус DONE

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/api/v1/tasks/1/status" `
-Method PATCH `
-ContentType "application/json" `
-Body '{"status":"done"}'
```

---

## Добавление комментария

(создайте новую задачу, которая ещё не находится в статусе DONE)

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/api/v1/tasks/2/comments" `
-Method POST `
-ContentType "application/json" `
-Body '{"author_id":1,"text":"Looks good"}'
```

---

## Получение комментариев задачи

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/api/v1/tasks/2/comments" `
-Method GET
```

---

## Закрытие проекта

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/api/v1/projects/1/close" `
-Method PATCH
```

---

## Проверка бизнес-правила: создание задачи в закрытом проекте

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/api/v1/tasks" `
-Method POST `
-ContentType "application/json" `
-Body '{"project_id":1,"assignee_id":1,"title":"Should fail","description":"Task in closed project"}'
```

Ожидаемый результат:

```json
{
  "error": "project is closed"
}
```

---

## Удаление задачи

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/api/v1/tasks/2" `
-Method DELETE
```

---

## Проверка удаления

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/api/v1/tasks/2" `
-Method GET
```

Ожидаемый результат:

```json
{
  "error": "resource not found"
}
```



## Бизнес-правила

* нельзя создать задачу в закрытом проекте;
* нельзя назначить задачу пользователю, который не является участником проекта;
* нельзя напрямую перевести задачу из `new` в `done`;
* из `done` и `cancelled` задача не переводится в другие статусы;
* нельзя добавить комментарий к задаче в статусе `done` или `cancelled`;
* повторный email пользователя возвращает ошибку `409 Conflict`.

