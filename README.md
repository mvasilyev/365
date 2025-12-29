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

## Local Development

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

### 2. Configure WebAuthn Origins

**Critical Step**: You must update the allowed origins for WebAuthn to match your public domain.

Edit `cmd/server/main.go` (around line 45) and `internal/auth/webauthn.go`:

**cmd/server/main.go**:
```go
authService, err := auth.NewService(
    db, 
    "photos.yourdomain.com", // <--- UPDATE THIS
    "https://photos.yourdomain.com", // <--- UPDATE THIS
    "365 Photos",
)
```

**internal/auth/webauthn.go**:
Update `RPOrigins` to include your domain:
```go
RPOrigins: []string{
    "https://photos.yourdomain.com", // <--- ADD YOUR DOMAIN
    "http://localhost:8080",
},
```

Rebuild the server after these changes.

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
