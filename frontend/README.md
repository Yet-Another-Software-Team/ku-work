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
GOOGLE_CLIENT_ID=your_google_client_id_here
TURNSTILE_CLIENT_TOKEN=your_turnstile_client_token
```

### 3. E2E Tests
#### 3.1 Playwright codegen
```
bunx playwright codegen your_frontend_base_url
```

#### 3.2 Run playwright
3.3.1 Run e2e test with normal playwright (config headless at [`playwright.config.ts`](./playwright.config.ts))
```
bun test:e2e
```

3.3.2 Run e2e test with playwright ui
```
bun test:e2e:ui
```

## Development

### Start Development Server

```bash
bun run dev
```

The application will be available at `http://localhost:3000`

### Available Scripts

| Command            | Description                             |
| ------------------ | --------------------------------------- |
| `bun run dev`      | Start development server                |
| `bun run build`    | Build for production                    |
| `bun run preview`  | Preview production build                |
| `bun run generate` | Generate static site                    |
| `bun lint`         | Check project file with eslint          |
| `bun lint:fix`     | Fix project file formatting with eslint |
| `bun format`       | Check project file format with prettier |
| `bun format:fix`   | Fix project file format with prettier   |
| `bun test:e2e`     | Run e2e test using playwright           |

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
