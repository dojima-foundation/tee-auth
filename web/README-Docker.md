# Web Frontend Docker Setup

This Docker setup contains only the web frontend service, designed to connect to external backend services.

## Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- External gauth backend service running (see `../gauth/` directory)

## Quick Start

1. **Copy environment variables:**
   ```bash
   cp .env.example .env
   ```

2. **Update environment variables:**
   Edit `.env` file with your actual values:
   - `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` for OAuth
   - `NEXTAUTH_SECRET` for authentication
   - `NEXT_PUBLIC_API_URL` and `NEXT_PUBLIC_GRPC_URL` to point to your backend

3. **Start the web service:**
   ```bash
   docker compose up -d
   ```

4. **Access the application:**
   - Web Frontend: http://localhost:3000
   - Health Check: http://localhost:3000/api/health

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NODE_ENV` | Node environment | `production` |
| `NEXT_PUBLIC_API_URL` | API backend URL | `http://localhost:8083` |
| `NEXT_PUBLIC_GRPC_URL` | gRPC backend URL | `http://localhost:9091` |
| `NEXTAUTH_URL` | NextAuth.js URL | `http://localhost:3000` |
| `NEXTAUTH_SECRET` | NextAuth.js secret | `development-nextauth-secret-change-in-production` |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | - |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret | - |
| `LOG_LEVEL` | Logging level | `info` |

## Backend Integration

This web service is designed to connect to external backend services. Make sure your backend services are running:

1. **Start the gauth backend:**
   ```bash
   cd ../gauth
   docker compose up -d
   ```

2. **Verify backend connectivity:**
   - API: http://localhost:8083
   - gRPC: http://localhost:9091

## Development

### Local Development
```bash
# Start the web service
docker compose up web

# View logs
docker compose logs -f web

# Stop the service
docker compose down
```

### Building the Image
```bash
# Build the web image
docker compose build web

# Build without cache
docker compose build --no-cache web
```

## Health Checks

The application includes a health check endpoint:
- **Endpoint**: `GET /api/health`
- **Response**: JSON with status, timestamp, uptime, and environment info

## Troubleshooting

### Common Issues

1. **Backend connection issues:**
   - Verify backend services are running
   - Check `NEXT_PUBLIC_API_URL` and `NEXT_PUBLIC_GRPC_URL` in `.env`
   - Ensure no firewall blocking connections

2. **OAuth issues:**
   - Verify Google OAuth credentials in `.env`
   - Ensure redirect URIs are configured in Google Console

3. **Build failures:**
   - Ensure Docker has enough memory (2GB+ recommended)
   - Clear Docker cache: `docker system prune -a`

### Debugging

```bash
# Check service status
docker compose ps

# Check service health
docker compose exec web wget -qO- http://localhost:3000/api/health

# Access service shell
docker compose exec web sh

# View service logs
docker compose logs -f web
```

## Architecture

```
┌─────────────────┐    ┌─────────────────┐
│   Web Frontend  │────│  Backend API    │
│   (Port 3000)   │    │  (Port 8083)    │
└─────────────────┘    └─────────────────┘
                                │
                       ┌─────────────────┐
                       │  gRPC Service   │
                       │  (Port 9091)    │
                       └─────────────────┘
```

The web frontend is a standalone service that communicates with external backend services via HTTP and gRPC.
