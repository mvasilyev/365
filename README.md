# 365 Photo Project

A minimal, self-hosted daily photo journaling application. Built with Go (backend) and React (frontend).

## Features
- **Daily Calendar**: Visualize your year in photos.
- **WebAuthn**: Passwordless login (Passkeys).
- **EXIF Data**: Auto-displays camera settings and location.
- **Map View**: Integrated OpenStreetMap for geotagged photos.
- **Theme**: Light, Dark, and System modes.

## Prerequisites
- **Go** (1.22+)
- **Node.js** (20+)
- **Caddy** (for HTTPS/Proxy)

## Quick Start (Automated)

Run the installation script to configure, build, and setup Caddy automatically:

```bash
chmod +x install.sh
./install.sh
```

## Manual Installation

If you prefer to set up manually:

### 1. Local Development

1.  **Backend**:
    ```bash
    go mod download
    go run ./cmd/server
    ```
    Runs on `:8080`.

2.  **Frontend**:
    ```bash
    cd client
    npm install
    npm run dev
    ```
    Runs on `:5173`.

3.  **HTTPS (Required for WebAuthn)**:
    Since WebAuthn requires a secure context (HTTPS) or `localhost`, accessing via a network IP requires a proxy.
    ```bash
    # Generate certs (if needed, or use Caddy's auto-internal)
    go run ./cmd/gen-cert
    
    # Run Caddy
    sudo caddy run
    ```

## Custom Domain Deployment

To host this on your own server (e.g., VPS) with a custom domain (e.g., `photos.yourdomain.com`).

### 1. Build the Application

Build the frontend (SPA) and backend binary:

```bash
# 1. Build Client
cd client
npm install
npm run build
cd ..

# 2. Build Server
go build -o server ./cmd/server
```

### 2. Configure Environment

Copy the example environment file:
```bash
cp .env.example .env
```

Edit `.env` to match your domain:
```ini
APP_DOMAIN=photos.yourdomain.com
APP_ORIGIN=https://photos.yourdomain.com
```

The server will automatically load these values.

### 3. Setup Caddy (Reverse Proxy)

Caddy handles SSL certificates automatically. Create a `Caddyfile` in the root (or `/etc/caddy/Caddyfile`):

```caddy
photos.yourdomain.com {
    reverse_proxy localhost:8080
}
```

Run Caddy:
```bash
caddy run
```

### 4. Run the Server

Ensure the `server` binary, `client/dist` folder, and `uploads` folder are in the same directory (or adjust paths).

```bash
./server
```

Your app should now be live at `https://photos.yourdomain.com`.

## Security Note

- **First Run**: The first user to register becomes the admin. Registration is automatically closed afterwards.
- **Backups**: Backup `photos.db` and the `uploads/` directory regularly.
