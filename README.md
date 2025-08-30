# KU-Work

A full-stack web application with Go backend and Nuxt.js frontend, featuring user management with PostgreSQL database.

## Prerequisites

### Required Software
- **Go** 1.23.3 or higher
- **Node.js** 18+ and **Bun**
- **PostgreSQL** database
- **Git**

### Optional (Recommended)
- **Docker** and **Docker Compose** for containerized development

## Installation (Devlelopment)

### 1. Clone the Repository

### 2. Backend Setup

#### Navigate to Backend Directory
```bash
cd backend
```

#### Install Go Dependencies
```bash
go mod download
```

#### Environment Configuration
1. Copy the sample environment file:
   ```bash
   # Linux/Unix/MacOS
   cp sample.env .env
   
   # Windows
   copy sample.env .env
   
   # Edit .env with your configuration
   ```

2. Edit `.env` file with your database credentials:
   ```env
   DB_USERNAME=your_db_username
   DB_PASSWORD=your_db_password
   DB_HOST=localhost
   DB_PORT=5432
   DB_NAME=ku_work_db
   LISTEN_ADDRESS=:8000
   
   # CORS Configuration (optional)
   CORS_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000
   CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
   CORS_ALLOWED_HEADERS=Origin,Content-Length,Content-Type,Authorization
   CORS_ALLOW_CREDENTIALS=false
   
   # JWT Configuration
   JWT_SECRET=your_jwt_secret_here
   
   # Google OAuth Configuration
   GOOGLE_CLIENT_ID=your_google_client_id_here
   GOOGLE_CLIENT_SECRET=your_google_client_secret_here
   
   ```

### 3. Frontend Setup

#### Navigate to Frontend Directory
```bash
cd ../frontend
```

#### Install Dependencies
Using Bun:
```bash
bun install
```

#### Environment Configuration
1. Copy the sample environment file:
   ```bash
   # Linux/Unix/MacOS
   cp sample.env .env
   
   # Windows
   copy sample.env .env
   
   # Edit .env with your configuration
   ```

2. Edit `.env` file with your backend URL (default should work for local development):
   ```env
   API_BASE_URL=http://localhost:8000
   GOOGLE_CLIENT_ID=your_google_client_id_here
   ```

## How to Run

### Option 1: Using Docker Compose (Recommended)

1. **Start all services**:
   ```bash
   cd backend
   docker compose up -d
   ```
   and
   ```bash
   cd frontend
   bun run dev
   ```

2. **Access the application**:
   - Backend API: http://localhost:8000
   - Frontend: http://localhost:3000

### Option 2: Manual Setup

#### 1. Start PostgreSQL Database
Make sure PostgreSQL is running and create a database:
```sql
CREATE DATABASE ku_work_db;
```
replace `ku_work_db` with your desired database name

#### 2. Start Backend Server
```bash
cd backend
go run .
```
The backend will start on **http://localhost:8000**

#### 3. Start Frontend Development Server
In a new terminal:
```bash
cd frontend
bun run dev
```
The frontend will start on **http://localhost:3000**

### Access Frontend
Open your browser and navigate to `http://localhost:3000`

## Development Commands

### Backend
```bash
cd backend
go run main.go          # Run development server
go build               # Build executable
```

### Frontend
```bash
cd frontend
bun run dev            # Start development server
bun run build          # Build for production
bun run preview        # Preview production build
```


---

**Built with Go, Nuxt, and PostgreSQL**