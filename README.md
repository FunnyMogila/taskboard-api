# Go chi demo server

Учебный пример веб-сервера на Go с использованием `chi`. Проект показывает:

- базовый HTTP server setup;
- middleware (`RequestID`, logging, recovery, timeout, real IP);
- route groups и versioned API;
- JSON CRUD без БД, на in-memory store;
- JWT login и protected routes;
- role-based authorization;
- query params, path params и headers;
- работу с `application/x-www-form-urlencoded`;
- multipart file upload;
- раздачу статических файлов;
- graceful shutdown.

## Run

```bash
go mod tidy
go run ./cmd/server
```

Переменные окружения:

- `PORT` default: `8080`
- `JWT_SECRET` default: `dev-secret-change-me`

## Demo users

- `alice` / `wonderland` -> role `admin`
- `bob` / `builder` -> role `user`

## Endpoints

- `GET /health`
- `GET /ready`
- `GET /static/`
- `GET /request-info`
- `POST /login`
- `POST /submit-form`
- `GET /api/v1/items`
- `GET /api/v1/items/{itemID}`
- `POST /api/v1/items`
- `PUT /api/v1/items/{itemID}`
- `PATCH /api/v1/items/{itemID}`
- `DELETE /api/v1/items/{itemID}`
- `GET /me`
- `POST /upload`
- `GET /admin/audit`

## Example requests

```bash
curl http://localhost:8080/health
```

```bash
curl http://localhost:8080/api/v1/items?page=1&page_size=5&status=published
```

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"wonderland"}'
```

```bash
curl -X POST http://localhost:8080/submit-form \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "name=Alex&email=alex@example.com&comments=Hello+from+form"
```

```bash
curl -X POST http://localhost:8080/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@README.md"
```

```bash
curl -X POST http://localhost:8080/api/v1/items \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Write docs",
    "description": "Document the chi example server",
    "status": "draft",
    "tags": ["docs", "demo"]
  }'
```
