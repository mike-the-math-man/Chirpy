# Chirpy

A simple Twitter-like API built with Go as part of the boot.dev Go HTTP Servers course. This project demonstrates building a REST API with user management, chirp creation, profanity filtering, and basic admin functionality.

## Features

- **User Management**: Create users with email addresses.
- **Chirp Creation**: Post short messages (chirps) with automatic profanity filtering.
- **Admin Tools**: View metrics and reset data (in dev mode).
- **File Serving**: Static file serving for web assets.
- **Health Check**: Basic API health endpoint.

## API Endpoints

- `GET /api/healthz` - Health check
- `GET /admin/metrics` - View server metrics (hit count)
- `POST /admin/reset` - Reset metrics and delete all users (dev mode only)
- `POST /api/users` - Create a new user
- `POST /api/chirps` - Create a new chirp
- `GET /app/*` - Serve static files

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
