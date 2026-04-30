# Chirpy

A simple Twitter-like API built with Go as part of the boot.dev Go HTTP Servers course. This project demonstrates building a REST API with user management, chirp creation, profanity filtering, and basic admin functionality.

## Features

- **User Management**: Create users with email addresses.
- **Chirp Creation**: Post short messages (chirps) with automatic profanity filtering.
- **Admin Tools**: View metrics and reset data (in dev mode).
- **File Serving**: Static file serving for web assets.
- **Health Check**: Basic API health endpoint.

## API Endpoints

The API exposes the following resources, paths, HTTP methods, and JSON shapes.

### Public Endpoints

- `GET /api/healthz`
  - Description: Health check
  - Request: none
  - Response: plain text `OK`

- `POST /api/users`
  - Description: Create a new user
  - Request JSON:
    - `email` (string)
    - `password` (string)
  - Response JSON:
    - `id` (UUID)
    - `created_at` (timestamp)
    - `updated_at` (timestamp)
    - `email` (string)

- `POST /api/login`
  - Description: Authenticate a user
  - Request JSON:
    - `email` (string)
    - `password` (string)
  - Response JSON:
    - `id` (UUID)
    - `created_at` (timestamp)
    - `updated_at` (timestamp)
    - `email` (string)
    - `token` (JWT string)
    - `refresh_token` (string)

### Authenticated User Endpoints

- `PUT /api/users`
  - Description: Update the authenticated user's email and password
  - Requires: `Authorization: Bearer <JWT>` header
  - Request JSON:
    - `email` (string)
    - `password` (string)
  - Response JSON:
    - `id` (UUID)
    - `created_at` (timestamp)
    - `updated_at` (timestamp)
    - `email` (string)

- `POST /api/chirps`
  - Description: Create a new chirp
  - Requires: `Authorization: Bearer <JWT>` header
  - Request JSON:
    - `body` (string, max 140 characters)
    - `user_id` (UUID) — accepted but ignored; the authenticated user is used instead
  - Response JSON:
    - `id` (UUID)
    - `created_at` (timestamp)
    - `updated_at` (timestamp)
    - `body` (string)
    - `user_id` (UUID)

- `GET /api/chirps`
  - Description: List all chirps
  - Request: none
  - Response JSON: array of chirp objects
    - `id` (UUID)
    - `created_at` (timestamp)
    - `updated_at` (timestamp)
    - `body` (string)
    - `user_id` (UUID)

- `GET /api/chirps/{chirpID}`
  - Description: Get a single chirp by its ID
  - Request: none
  - Response JSON:
    - `id` (UUID)
    - `created_at` (timestamp)
    - `updated_at` (timestamp)
    - `body` (string)
    - `user_id` (UUID)

### Refresh Token Endpoints

- `POST /api/refresh`
  - Description: Exchange a refresh token for a new JWT
  - Requires: `Authorization: Bearer <refresh_token>` header
  - Response JSON:
    - `token` (string)

- `POST /api/revoke`
  - Description: Revoke the current refresh token
  - Requires: `Authorization: Bearer <refresh_token>` header
  - Response: HTTP 204 No Content

### Admin / Dev Endpoints

- `GET /admin/metrics`
  - Description: View server metrics / hit count
  - Request: none
  - Response: HTML page showing visit count

- `POST /admin/reset`
  - Description: Reset metrics and delete all users
  - Requires: `PLATFORM=dev`
  - Request: none
  - Response: plain text confirmation

### Static Files

- `GET /app/*`
  - Description: Serve static assets from the project root
  - Example: `GET /app/index.html`

## Dependencies

- Go 1.19 or later
- PostgreSQL database
- Required Go modules (automatically installed via `go mod tidy`):
  - `github.com/pressly/goose/v3` - Database migrations
  - `github.com/joho/godotenv` - Environment variable loading
  - `github.com/lib/pq` - PostgreSQL driver
  - `github.com/google/uuid` - UUID generation

## Setup and Installation

### 1. Install Dependencies

- **Go**: Download from [golang.org](https://golang.org/dl/)
- **PostgreSQL**: Install via your package manager or download from [postgresql.org](https://www.postgresql.org/download/)

### 2. Clone and Setup

```bash
git clone <repository-url>
cd Chirpy
go mod tidy
```

### 3. Database Setup

Create a PostgreSQL database named `chirpy`:

```sql
CREATE DATABASE chirpy;
```

Update the `.env` file with your database connection (default provided):

### 4. Run Database Migrations

Use Goose to apply the database schema:

```bash
goose -dir sql/schema postgres "$DB_URL" up
```

To reset the database (rollback all migrations):

```bash
goose -dir sql/schema postgres "$DB_URL" reset
```

### 5. Generate Database Code

Generate Go code from SQL queries using sqlc:

```bash
sqlc generate
```

### 6. Run the Server

```bash
go run .
```

The server will start on `http://localhost:8080`.

## Usage

### Creating a User

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com"}'
```

### Creating a Chirp

```bash
curl -X POST http://localhost:8080/api/chirps \
  -H "Content-Type: application/json" \
  -d '{"body": "Hello, world!", "user_id": "your-user-uuid"}'
```

Note: Chirps are limited to 140 characters and profanity words ("kerfuffle", "sharbert", "fornax") are filtered.

### Admin Reset (Dev Mode Only)

```bash
curl -X POST http://localhost:8080/admin/reset
```

This resets the hit counter and deletes all users.

## Building as Executable

To build a standalone executable:

```bash
go build -o chirpy .
```

### Running the Executable

1. Ensure PostgreSQL is running and the database is set up (steps 3-5 above).
2. Set environment variables or use a `.env` file in the same directory as the executable:
   ```
   DB_URL="postgres://postgres:postgres@localhost:5432/chirpy?sslmode=disable"
   PLATFORM="dev"  # or "prod" for production
   ```
3. Run the executable:
   ```bash
   ./chirpy
   ```

The server will start on `http://localhost:8080`.

### Configuration

- **DB_URL**: PostgreSQL connection string
- **PLATFORM**: Set to "dev" for development features (like reset endpoint), "prod" for production

## Project Structure

- `main.go` - Main server code
- `internal/database/` - Generated database code (sqlc)
- `sql/schema/` - Database migration files (Goose)
- `sql/queries/` - SQL query files (sqlc)
- `assets/` - Static web assets
- `index.html` - Main HTML file

## Development

- Use `go mod tidy` to manage dependencies
- Run `sqlc generate` after modifying SQL queries
- Use Goose commands for database schema changes
