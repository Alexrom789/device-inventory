# Device Inventory Service

A Go microservice for managing device lifecycle in a refurbishment/resale workflow. Built with Fiber, PostgreSQL, and Go's native concurrency primitives.

## Architecture

```
Client (Postman / curl)
        ↓
Fiber HTTP Layer  (handlers/)
        ↓
Service Layer     (service/)      ← business logic + goroutines/channels
        ↓
Repository Layer  (repository/)   ← all SQL queries
        ↓
PostgreSQL
```

Each layer has one responsibility and only talks to the layer directly below it. This makes the codebase testable, readable, and easy to extend.

## Features

- Full CRUD for device inventory
- Device lifecycle status transitions: `received → testing → graded → sold`
- **Async device processing** using Go goroutines and channels
- Input validation and structured error responses
- Connection pooling via `sqlx`
- Request logging middleware

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.21 |
| HTTP Framework | Fiber v2 |
| Database | PostgreSQL |
| DB Driver | sqlx + lib/pq |
| Config | godotenv |

## Project Structure

```
device-inventory/
├── cmd/
│   └── main.go                  # Entry point, dependency wiring
├── config/
│   └── database.go              # DB connection + pool config
├── internal/
│   ├── handlers/
│   │   └── device_handler.go    # HTTP request/response handling
│   ├── models/
│   │   └── device.go            # Domain types, request structs
│   ├── repository/
│   │   └── device_repository.go # SQL queries (Postgres)
│   └── service/
│       └── device_service.go    # Business logic, goroutines
├── db/
│   └── migrations/
│       └── 001_create_devices.sql
├── .env
├── .gitignore
├── go.mod
└── README.md
```

## Getting Started

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [PostgreSQL 14+](https://www.postgresql.org/download/)

### 1. Clone the repo

```bash
git clone https://github.com/alexrom789/device-inventory.git
cd device-inventory
```

### 2. Set up the database

Open `psql` and run:

```sql
CREATE DATABASE device_inventory;
\c device_inventory
\i db/migrations/001_create_devices.sql
```

### 3. Configure environment

Edit `.env` to match your Postgres credentials:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=device_inventory
PORT=3000
```

### 4. Install dependencies & run

```bash
go mod tidy
go run cmd/main.go
```

You should see:
```
Database connected successfully
Server starting on port 3000
```

## API Reference

### Health Check
```
GET /health
```

### Create a Device
```
POST /devices
Content-Type: application/json

{
  "imei": "123456789012345",
  "model": "iPhone 13 Pro",
  "price": 499.99
}
```

### Get All Devices
```
GET /devices
```

### Get Device by ID
```
GET /devices/:id
```

### Update Status
```
PUT /devices/:id/status
Content-Type: application/json

{
  "status": "testing"
}
```
Valid statuses: `received`, `testing`, `graded`, `sold`

### Process Device (Async Grading)
```
POST /devices/:id/process
```

Launches a goroutine to simulate warehouse testing. The device status is set to `testing` immediately, then graded asynchronously. Returns the grade result when processing completes.

Example response:
```json
{
  "device_id": "abc-123",
  "new_grade": "A",
  "message": "Device iPhone 13 Pro graded as A after 2s of testing"
}
```

## Concurrency Design

The `/process` endpoint demonstrates Go's concurrency model:

1. A **goroutine** runs `simulateGrading` concurrently — cheap to create, managed by the Go runtime
2. A **channel** (`resultChan`) safely transfers the result back — no shared memory, no race conditions
3. A **select with timeout** ensures the request never hangs indefinitely

```go
resultChan := make(chan models.ProcessResult, 1)
go s.simulateGrading(device, resultChan)

select {
case result := <-resultChan:
    // success
case <-time.After(10 * time.Second):
    // timeout
}
```

This pattern is idiomatic Go and maps directly to real-world use cases like async job processing, webhook delivery, and background workers.

UI SAMPLE:
<img width="2021" height="1324" alt="GoAppUI" src="https://github.com/user-attachments/assets/d4b043c1-f61f-4521-935f-5af455f0a9af" />
