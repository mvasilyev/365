#!/bin/bash
set -e

echo "📸 365 Photo Project Installer"
echo "=============================="

# 1. Check Dependencies
echo "Checking dependencies..."
if ! command -v go &> /dev/null; then echo "❌ Go is not installed"; exit 1; fi
if ! command -v npm &> /dev/null; then echo "❌ Node.js (npm) is not installed"; exit 1; fi

# 2. Configuration
echo ""
echo "--- Configuration ---"
if [ -f .env ]; then
    echo "Found existing .env file."
    source .env
    DOMAIN=$APP_DOMAIN
else
    read -p "Enter your domain (e.g. photos.example.com or 192.168.x.x.sslip.io): " DOMAIN
    if [ -z "$DOMAIN" ]; then echo "Domain required"; exit 1; fi
    
    ORIGIN="https://$DOMAIN"
    echo "Using Origin: $ORIGIN"
    
    echo "APP_DOMAIN=$DOMAIN" > .env
    echo "APP_ORIGIN=$ORIGIN" >> .env
    echo "✅ Created .env"
fi

# 3. Build
echo ""
echo "--- Building ---"

echo "📦 Building Client..."
cd client
npm install
npm run build
cd ..

echo "📦 Building Server..."
go build -o server ./cmd/server

# 4. Caddy
echo ""
echo "--- Caddy Setup ---"
if command -v caddy &> /dev/null; then
    cat > Caddyfile <<EOF
$DOMAIN {
    reverse_proxy localhost:8080
}
EOF
    echo "✅ Generated Caddyfile for $DOMAIN"
    
    echo ""
    echo "To run Caddy, you can:"
    echo "1. Run in foreground: sudo caddy run"
    echo "2. Start in background: sudo caddy start"
    echo "3. Reload if running: sudo caddy reload"
    
    echo ""
    read -p "Do you want to starting Caddy now (foreground)? [y/N] " RUN_CADDY
    if [[ "$RUN_CADDY" =~ ^[Yy]$ ]]; then
        echo "Starting Caddy (Ctrl+C to stop)..."
        sudo caddy run
    fi
else
    echo "⚠️  Caddy not found. Please install Caddy to serve HTTPS."
    echo "Generated config would be:"
    echo "$DOMAIN {"
    echo "    reverse_proxy localhost:8080"
    echo "}"
fi

echo ""
echo "🎉 Setup Complete!"
echo "Run the server with: ./server"
