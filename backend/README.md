# Backend API

A RESTful API built with Go, Gin framework, PostgreSQL, and GORM for user management operations.

## Prerequisites

- Go 1.23.3 or higher
- PostgreSQL 13+ (or Docker for containerized setup)

## Installation

1. **Install Go dependencies**
   ```bash
   go mod download
   ```

2. **Set up environment variables**
   ```bash
   # Linux/Unix/MacOS
   cp sample.env .env

   # Windows
   copy sample.env .env

   # Edit .env with your configuration
   ```

3. **Set up the database**
   - Option A: Use Docker Compose (recommended)
     ```bash
     docker compose up -d db
     ```
   - Option B: Install PostgreSQL locally and create a database
   ```sql
    CREATE DATABASE ku_work_db;
    ```
    replace `ku_work_db` with your desired database name


## Configuration

Copy `sample.env` to `.env` and configure the following variables:

### Database Configuration
- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5432)
- `DB_USERNAME`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name

### Server Configuration
- `LISTEN_ADDRESS`: Server listen address (default: :8080)

### CORS Configuration
- `CORS_ALLOWED_ORIGINS`: Comma-separated list of allowed origins
- `CORS_ALLOWED_METHODS`: Allowed HTTP methods
- `CORS_ALLOWED_HEADERS`: Allowed headers
- `CORS_ALLOW_CREDENTIALS`: Allow credentials (true/false)
- `CORS_MAX_AGE`: Preflight cache duration in seconds

### JWT Configuration
- `JWT_SECRET`: Secret key for JWT token generation

### Cookie Configuration
- `COOKIE_SECURE`: Enable secure flag for cookies (true/false, default: true)
  - Set to `true` for HTTPS environments (production)
  - Set to `false` for HTTP environments (local development)
  - Note: Cookies use SameSite=None for cross-origin support

### GoogleOauth Configuration
- `GOOGLE_CLIENT_ID`: Client ID for Google OAuth
- `GOOGLE_CLIENT_SECRET`: Client secret for Google OAuth

### Swagger Configuration
- `SWAGGER_HOST`: Swagger host (default: localhost:8000)

### AI Configuration
- `APPROVAL_AI`: Choose what AI to use (dummy, ollama, ...)

**Ollama Configuration**
- `APPROVAL_AI_MODEL`: Choose what AI model to use (e.g. gemma3)
- `APPROVAL_AI_URI`: Endpoint of AI server

### Email Configuration
- `EMAIL_PROVIDER`: Choose what email provider to use (dummy, SMTP, gmail, ...)

If you use other provider than dummy follow the [configuration guide](./email_config.md) here.

## Running the Application

### Development Mode

1. **Start the database** (if using Docker)
   ```bash
   docker compose up -d db
   ```

2. **Run the application**
   ```bash
   go run main.go
   ```

### Production Build (WIP, NOT TESTED)

1. **Build the binary**
   ```bash
   go build -o bin/api main.go
   ```

2. **Run the binary**
   ```bash
   ./bin/api
   ```

## Docker Compose Usage

The project includes Docker Compose configuration for easy database usage:

1. **Start the backend services**
   ```bash
   docker compose up -d
   ```

2. **Stop the server**
   ```bash
   docker compose down
   ```

3. **Reset server (remove volumes)**
   ```bash
   docker compose down -v
   ```

## Creating an Admin User

You can create an admin user by running a script. You must provide a username, and you will be prompted to enter a password securely.

### Using Docker

To run the script inside a running Docker container, you must use the `-it` flags to enable an interactive terminal for the password prompt.

```bash
docker compose exec -it <container> /app/create_admin <username>
```

- Replace `<container>` with the name of the container you want to run the script in.
- Replace `<username>` with your desired username.

### Locally (requires Go)

If you have Go installed on your machine, you can run the script directly:

```bash
go run ./scripts/create_admin.go <username>
```

This will create a new user with admin privileges in the database.

### Code Organization

The codebase follows a modular structure:

- **`database/`**: Database connection and configuration logic
- **`handlers/`**: HTTP request handlers and routing logic
- **`middleware/`**: HTTP middleware (CORS, authentication, etc.)
- **`model/`**: Database models and schema definitions
- **`main.go`**: Application entry point and server setup

### Database Migrations

GORM handles automatic migrations when the application starts. New model fields will be automatically added to the database schema.

### Swagger Documentation

Swagger documentation is available at `/swagger/index.html`.

[Swaggo Documentation](https://github.com/swaggo/gin-swagger)

To update the Swagger documentation, run the following command:

On linux
```bash
~/go/bin/swag init -g main.go
```

On windows - Does not tested
```bash
swag init -g main.go
```
