# KU-Work Frontend

A modern Vue.js frontend application built with Nuxt 3, providing a user-friendly interface for the KU-Work user management system.

## Prerequisites

- **Node.js** 18+
- **Bun**
- **Backend API** running on http://localhost:8000 (default)

## Installation

### 1. Install Dependencies

Using Bun:
```bash
bun install
```

### 2. Environment Configuration

Copy the sample environment file and configure your API settings:

   ```bash
   # Linux/Unix/MacOS
   cp sample.env .env

   # Windows
   copy sample.env .env

   # Edit .env with your configuration
   ```

Edit `.env` with your backend API URL:
```env
API_BASE_URL=http://localhost:8000
GOOGLE_CLIENT_SECRET=your_google_client_secret_here
```

## Development

### Start Development Server

```bash
bun run dev
```

The application will be available at `http://localhost:3000`

### Available Scripts

| Command | Description |
|---------|-------------|
| `bun run dev` | Start development server |
| `bun run build` | Build for production |
| `bun run preview` | Preview production build |
| `bun run generate` | Generate static site |

## Production Deployment

### Build Application

```bash
bun run build
```

### Preview Production Build

```bash
bun run preview
```

### Static Generation (Optional)

For static hosting:
```bash
bun run generate
```
