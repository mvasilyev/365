# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

365 Photo Project is a self-hosted daily photo journaling application. Go backend (Chi router, SQLite) + React/TypeScript frontend (Vite). Uses WebAuthn (Passkeys) for passwordless authentication.

## Commands

### Development

```bash
# Backend (runs on :8080)
go run ./cmd/server

# Frontend (runs on :5173)
cd client && npm install && npm run dev

# HTTPS required for WebAuthn - generate certs and run Caddy
go run ./cmd/gen-cert
sudo caddy run
```

### Production Build

```bash
# Build frontend
cd client && npm install && npm run build

# Build backend
go build -o server ./cmd/server
```

### Utilities

```bash
# Seed database with test data (45 days of photos)
go run ./cmd/seed

# Lint frontend
cd client && npm run lint
```

## Architecture

### Backend (`cmd/`, `internal/`)

- **cmd/server/main.go**: Entry point, HTTP server setup, routes configuration
- **internal/api/handler.go**: HTTP handlers for photos and authentication endpoints
- **internal/auth/webauthn.go**: WebAuthn service, user/credential management
- **internal/store/photos.go**: Photo persistence layer (SQLite)
- **internal/store/schema.sql**: Database schema

### Frontend (`client/src/`)

- **App.tsx**: Main routing and theme provider
- **api.ts**: API client with WebAuthn browser API integration
- **GalleryView.tsx**: Month-grouped photo calendar
- **DetailView.tsx**: Single photo view with map
- **UploadView.tsx**: Photo upload with notes
- **LoginView.tsx**: WebAuthn registration/login

### Data Flow

1. Photos uploaded → EXIF extracted (date, GPS, camera) → thumbnail generated (400x400) → stored in `uploads/`
2. Photo date from EXIF becomes the unique key (one photo per day)
3. Frontend served as SPA from `client/dist/`, API calls to `/api/*`

### Database (SQLite - `photos.db`)

- **users**: id, username, credentials (WebAuthn blob)
- **photos**: day (YYYY-MM-DD primary key), filepath, thumbnail_path, lat, lon, notes, exif_data
- **sessions**: token, user_id, expires_at (30-day TTL)

## Environment Configuration

Copy `.env.example` to `.env`:

```ini
APP_DOMAIN=localhost          # WebAuthn Relying Party ID
APP_ORIGIN=http://localhost:8080  # Full origin for WebAuthn
```

For production, use HTTPS domain (e.g., `APP_DOMAIN=photos.example.com`, `APP_ORIGIN=https://photos.example.com`).

## Key Constraints

- WebAuthn requires HTTPS or localhost (use Caddy for non-localhost development)
- First registered user closes registration for others
- Photo day is unique - uploading for same day replaces existing photo
