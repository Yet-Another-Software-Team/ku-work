# Backend API

A RESTful API built with Go, Gin framework, PostgreSQL, Redis, and GORM for user management and authentication operations.

## Prerequisites

- Go 1.23.3 or higher
- PostgreSQL 13+ (or Docker for containerized setup)
- Redis 7+ (or Docker for containerized setup)
- Optional: Google Cloud project (if using GCS provider)

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

Copy `sample.env` to `.env` and configure the following variables. The repository includes `backend/sample.env` with defaults and additional storage provider variables (`FILE_PROVIDER`, `LOCAL_FILES_DIR`, `GCS_*`) to select how files are stored.

### Database Configuration
- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5432)
- `DB_USERNAME`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name

### Redis Configuration
Redis is used for rate limiting and session management.

- `REDIS_HOST`: Redis host (default: localhost)
- `REDIS_PORT`: Redis port (default: 6379)
- `REDIS_PASSWORD`: Redis password (optional, leave empty if not required)
- `REDIS_DB`: Redis database number (default: 0)

**Note**: The application will start even if Redis is unavailable, but rate limiting will be disabled (fail-open behavior). For production deployments, ensure Redis is running for security features.

### Server Configuration
- `LISTEN_ADDRESS`: Server listen address (default: :8080)

### CORS Configuration
- `CORS_ALLOWED_ORIGINS`: Comma-separated list of allowed origins
- `CORS_ALLOWED_METHODS`: Allowed HTTP methods
- `CORS_ALLOWED_HEADERS`: Allowed headers
- `CORS_ALLOW_CREDENTIALS`: Allow credentials (true/false)
- `CORS_MAX_AGE`: Preflight cache duration in seconds

### JWT Configuration
- `JWT_SECRET`: Secret key for JWT token generation (minimum 32 bytes recommended)
  - JWTs expire in 15 minutes
  - Include unique JTI (JWT ID) for blacklist support
  - OWASP-compliant token revocation on logout

### Session Configuration
- `MAX_SESSIONS_PER_USER`: Maximum number of concurrent sessions per user (default: 10)
  - Users can be logged in on multiple devices simultaneously
  - When limit is reached, oldest sessions are automatically revoked
  - Set to 1 for single-session-only (most secure)
  - Higher values allow more convenience but increase token storage

### Cookie Configuration
- `COOKIE_SECURE`: Enable secure flag for cookies (true/false, default: true)
  - Set to `true` for HTTPS environments (production)
  - Set to `false` for HTTP environments (local development)
  - Note: Cookies use SameSite=None for cross-origin support

### GoogleOauth Configuration
- `GOOGLE_CLIENT_ID`: Client ID for Google OAuth
- `GOOGLE_CLIENT_SECRET`: Client secret for Google OAuth

### File storage provider
The backend offers a pluggable file storage provider. Configure the provider using the `FILE_PROVIDER` environment variable in `backend/.env`:

- `FILE_PROVIDER`: "local" (default) or "gcs"
- When using local storage:
  - `LOCAL_FILES_DIR` — directory where files are written (default `./files`)
- When using GCS:
  - `GCS_BUCKET` — the Google Cloud Storage Bucket name (i.e., my-bucket)
  - `GCS_CREDENTIALS_PATH` — path to Google cloud service-account JSON credentials on your machine.
  - > **Important:** When running with Docker and using the GCS provider, you must:
  > 1. Place your `gcs-key.json` file in the `backend` directory.
  > 2. In your `.env` file, set `GCS_CREDENTIALS_PATH` to `/app/gcs-key.json`.

### Swagger Configuration
- `SWAGGER_HOST`: Swagger host (default: localhost:8000)

### AI Configuration
- `APPROVAL_AI`: Choose what AI to use (dummy, ollama, ...)

**Ollama Configuration**
- `APPROVAL_AI_MODEL`: Choose what AI model to use (e.g. gemma3)
- `APPROVAL_AI_URI`: Endpoint of AI server

### Email Configuration
- `EMAIL_PROVIDER`: Choose what email provider to use (dummy, SMTP, gmail, ...)
- `EMAIL_TIMEOUT_SECONDS`: Specify the timeout duration of email sending attempt in seconds

**Email Retry Configuration**
- `EMAIL_RETRY_MAX_ATTEMPTS`: Maximum number of retry attempts for failed emails (default: 3)
- `EMAIL_RETRY_INTERVAL_MINUTES`: Minutes to wait before retrying failed emails (default: 30)
- `EMAIL_RETRY_MAX_AGE_HOURS`: Maximum age of emails to retry in hours (default: 24)

The email service automatically retries emails that fail with temporary errors (e.g., network issues, rate limits, timeouts). Each retry attempt is tracked in the database (`RetryCount` field), and emails that exceed the maximum retry attempts are marked as permanent failures.

If you use other provider than dummy follow the [configuration guide](./email_config.md) here.

## Running the Application

### Development Mode

1. **Start the database and Redis** (if using Docker)
   ```bash
   docker compose up -d db redis
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

The project includes Docker Compose configuration for easy database and Redis usage:

1. **Start the backend services**
   ```bash
   docker compose up -d
   ```
   This will start PostgreSQL, Redis, and the backend API.

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

**Included Models**:
- `User`, `Admin`, `Company`, `Student`: User management
- `RefreshToken`: Refresh token storage with Argon2id hashing
- `RevokedJWT`: JWT blacklist for logout functionality
- `Job`, `JobApplication`: Job posting and application management
- `File`: File upload management
- `Audit`: Audit logging
- `GoogleOAuthDetails`: OAuth integration

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

## Monitoring & Maintenance

### Background Tasks

The application runs scheduled background tasks:

1. **Token Cleanup** (hourly)
   - Removes expired refresh tokens
   - Keeps revoked tokens for 7 days (reuse detection)

2. **JWT Blacklist Cleanup** (hourly)
   - Removes expired JWTs from blacklist
   - Prevents unbounded table growth

3. **Email Retry** (configurable)
   - Automatically retries emails that failed with temporary errors
   - Tracks actual retry attempts in database (`RetryCount` field)
   - Respects configured retry interval and max attempts
   - Marks emails as permanent failures after max attempts exceeded
   - Only retries emails within the configured maximum age window

### Security Monitoring

Monitor these metrics for security:
- Blocked revoked token attempts (search logs for "SECURITY: Blocked revoked JWT")
- Rate limit violations
- Token reuse detection alerts
- Blacklist table size: `SELECT COUNT(*) FROM revoked_jwts`

### Testing

Run the test suite:
```bash
go test ./tests -v
```

Run specific test:
```bash
go test ./tests -v -run TestJWTBlacklist
```

## Troubleshooting

### Redis Connection Issues
If Redis is unavailable:
- Application will start but log warnings
- Rate limiting will be disabled (fail-open)
- Consider this acceptable for development, but fix for production

### JWT Issues
- **"Token has been revoked"**: User logged out, need to login again
- **"Invalid token"**: Token expired (15 min), use refresh token endpoint
- **Tokens not being blacklisted**: Check database connection and `revoked_jwts` table

### Database Issues
- Check connection settings in `.env`
- Verify PostgreSQL is running: `docker compose ps`
- Check logs: `docker compose logs db`
```

