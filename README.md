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

## How to run

### Clone the Repository

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

#### Run Docker Compose
```bash
docker compose up
```


### Alternatives
You can also run frontend and backend service seprately by following the specific service guide
- [frontend](./frontend/README.md)
- [backend](./backend/README.md)

### Creating an Admin User (Docker)

To create an admin user while the services are running, execute the following command. It requires an interactive terminal (`-it`) to securely prompt for a password.

```bash
docker-compose exec -it <container-name> /app/create_admin <username>
```

- Replace `<container-name>` with the name of the container you want to run the command in.
- Replace `<username>` with your desired username.

---

**Built with Go, Nuxt, and PostgreSQL**
