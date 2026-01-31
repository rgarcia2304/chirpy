# Chirpy 

Chirpy is a Twitter-style backend service
It exposes a RESTful API for creating and managing short posts (“chirps”), with user authentication, admin functionality, and persistent storage backed by PostgreSQL.

The project emphasizes production-grade backend fundamentals: database migrations, type-safe queries, authentication, and clean API design.

---

## Features

- User registration and authentication
- JWT based access and refresh tokens
- Create, read, and delete chirps
- Authorization (users can only delete their own chirps)
- Admin only endpoints and metrics
- PostgreSQL backed persistence
- Database migrations with Goose
- Typesafe SQL queries via sqlc

---

## Tech Stack

- **Language:** Go
- **Database:** PostgreSQL 15
- **Migrations:** Goose
- **Query Generation:** sqlc
- **HTTP Server:** net/http
- **Authentication:** JWT

---

## Prerequisites

- Go 1.20+
- PostgreSQL 15
- Goose
- sqlc

---

## Setup

### Clone the Repository

```bash
git clone https://github.com/rgarcia2304/chirpy.git
cd chirpy
```

### Create Database and set the connection string 
```export DATABASE_URL="postgres://user:password@localhost:5432/chirpy?sslmode=disable"```

### Run Migrations
```goose -dir sql/schema postgres "$DATABASE_URL" up```

### Generate SQL Code 
```sqlc generate```

### Build and Run 
```
go build
./chirpy
```

[The server runs on localhost/](http://localhost:8080)

### Get Health
GET /api/healthz

### Users & Auth

POST /api/users – Create a user

POST /api/login – Authenticate and receive tokens

POST /api/refresh – Refresh access token

### Authentication uses JWT access tokens with refresh tokens.
Protected endpoints require:

Authorization: Bearer <access_token>

### Chirps

GET /api/chirps – List all chirps

GET /api/chirps/{id} – Retrieve a chirp by ID

POST /api/chirps – Create a chirp (auth required)

DELETE /api/chirps/{id} – Delete a chirp (owner only)

### Admin

GET /admin/metrics – View internal service metrics

POST /admin/reset – Reset application state

Admin endpoints require elevated privileges.

### Project Structure
```
chirpy/
├── main.go
├── sql/
│   ├── schema/        
│   └── queries/       
├── internal/
│   ├── auth/
│   ├── database/
│   └── handlers/
├── sqlc.yaml
└── README.md
```
