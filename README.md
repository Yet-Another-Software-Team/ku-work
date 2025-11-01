# KU-Work

A full-stack web application with a Go backend and Nuxt.js frontend, featuring user management with a PostgreSQL database.

## Prerequisites

### Required Software
- **Go** 1.23.3 or higher
- **Node.js** 18+ and **Bun**
- **PostgreSQL** database
- **Git**

### Optional (Recommended)
- **Docker** and **Docker Compose** for containerized development

## How to run

The easiest way to run the application is by using Docker Compose.

### 1. Clone the Repository
```bash
git clone <repository-url>
cd <repository-name>
```

### 2. Configure Environment
Copy the sample environment file and edit it with your configuration.
```bash
cp sample.env .env
```
For more detailed information on the environment variables, please refer to the [backend README](./backend/README.md).

> **Important:** When running with Docker and using the GCS provider, you must:
  > 1. Place your `gcs-key.json` file in the `backend` directory.
  > 2. In your `.env` file, set `GCS_CREDENTIALS_PATH` to `/app/gcs-key.json`.

### 3. Run Docker Compose
```bash
docker compose up
```
This will start the backend, frontend, database, and Redis services.

## Alternatives
You can also run the frontend and backend services separately by following the specific service guides:
- [frontend](./frontend/README.md)
- [backend](./backend/README.md)

## Creating an Admin User (Docker)

To create an admin user while the services are running, execute the following command. It requires an interactive terminal (`-it`) to securely prompt for a password.

```bash
docker-compose exec -it <container-name> /app/create_admin <username>
```

- Replace `<container-name>` with the name of the container you want to run the command in.
- Replace `<username>` with your desired username.

---

**Built with Go, Nuxt, and PostgreSQL**
