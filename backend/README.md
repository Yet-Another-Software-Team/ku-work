# Backend API

A RESTful API built with Go, Gin framework, PostgreSQL, and GORM for user management operations.

## Prerequisites

- Go 1.23.3 or higher
- PostgreSQL 13+ (or Docker for containerized setup)

## Installation

1. **Install Go dependencies**
   ```bash
   go mod tidy
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
     docker-compose up -d db
     ```
   - Option B: Install PostgreSQL locally and create a database

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

## Running the Application

### Development Mode

1. **Start the database** (if using Docker)
   ```bash
   docker-compose up -d db
   ```

2. **Run the application**
   ```bash
   go run main.go
   ```

### Production Build

1. **Build the binary**
   ```bash
   go build -o bin/api main.go
   ```

2. **Run the binary**
   ```bash
   ./bin/api
   ```

## API Endpoints

### Health Check
- **GET** `/`
  - Description: Health check endpoint
  - Response: `{"message": "Simple response"}`

### User Management

#### Get All Users
- **GET** `/users`
  - Description: Retrieve all users

#### Create User
- **POST** `/create_user`
  - Description: Create a new user
  - Request Body:
    ```json
    {
      "user": "username",
      "password": "password"
    }
    ```
  - Response: `"OK ;)"`
  - Status Codes:
    - `200`: User created successfully
    - `400`: Invalid request body
    - `500`: Internal server error

## Docker Setup

The project includes Docker Compose configuration for easy database setup:

1. **Start the database**
   ```bash
   docker-compose up -d db
   ```

2. **View database logs**
   ```bash
   docker-compose logs db
   ```

3. **Stop the database**
   ```bash
   docker-compose down
   ```

4. **Reset database (remove volumes)**
   ```bash
   docker-compose down -v
   ```
  
### Code Organization

The codebase follows a modular structure:

- **`database/`**: Database connection and configuration logic
- **`handlers/`**: HTTP request handlers and routing logic
- **`middleware/`**: HTTP middleware (CORS, authentication, etc.)
- **`model/`**: Database models and schema definitions
- **`main.go`**: Application entry point and server setup

### Database Migrations

GORM handles automatic migrations when the application starts. New model fields will be automatically added to the database schema.