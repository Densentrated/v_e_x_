# Podman Compose Setup

This document explains how to use podman-compose to run the backend server and Qdrant vector database in a containerized environment.

## Architecture

- **Backend**: Rust server (Actix Web) exposed on port 10801
- **Qdrant**: Vector database running internally (not exposed externally)
- **Network**: Both services communicate via internal `app-network`

## Prerequisites

- Podman installed
- podman-compose installed

## Quick Start

### Development Mode
```bash
# Navigate to project root
cd v_e_x_

# Start services in development mode
podman-compose --profile dev --env-file .env.dev up

# Start in background
podman-compose --profile dev --env-file .env.dev up -d

# View logs
podman-compose --profile dev logs -f
```

### Production Mode
```bash
# Navigate to project root
cd v_e_x_

# Start services in production mode
podman-compose --profile prod --env-file .env.prod up

# Start in background
podman-compose --profile prod --env-file .env.prod up -d
```

## Profiles

### Development Profile (`dev`)
- Uses `cargo watch` for hot reloading
- Mounts source code as volume for live editing
- Debug-level logging (`RUST_LOG=debug`)
- Faster build times with cached dependencies

### Production Profile (`prod`)
- Builds optimized release binary
- Info-level logging (`RUST_LOG=info`)
- No source code mounting
- Smaller container size

## Services

### Backend Service
- **Container Name**: `rust-backend`
- **External Port**: 10801
- **Internal Port**: 8080
- **Environment Variables**:
  - `RUST_LOG`: Logging level (debug/info)
  - `QDRANT_URL`: http://qdrant:6333

### Qdrant Service
- **Container Name**: `qdrant-server`
- **Internal Ports**: 6333 (REST), 6334 (gRPC)
- **External Access**: None (internal only)
- **Data Persistence**: `qdrant_data` volume

## Useful Commands

### Managing Services
```bash
# Stop all services
podman-compose --profile dev down

# Rebuild and start
podman-compose --profile dev up --build

# Force recreate containers
podman-compose --profile dev up --force-recreate
```

### Debugging
```bash
# Access backend container shell
podman-compose --profile dev exec backend /bin/bash

# View specific service logs
podman-compose --profile dev logs backend
podman-compose --profile dev logs qdrant

# Follow logs in real-time
podman-compose --profile dev logs -f backend
```

### Data Management
```bash
# List volumes
podman volume ls

# Remove all volumes (WARNING: destroys data)
podman-compose --profile dev down -v

# Backup Qdrant data
podman run --rm -v v_e_x__qdrant_data:/data -v $(pwd):/backup alpine tar czf /backup/qdrant-backup.tar.gz -C /data .
```

## Connecting to Qdrant

From your Rust application, connect to Qdrant using:
- **REST API**: `http://qdrant:6333`
- **gRPC API**: `qdrant:6334`

The `QDRANT_URL` environment variable is automatically set in the container.

## Network Security

- Qdrant is **not exposed** to external network
- Only the backend service is accessible from outside (port 10801)
- All inter-service communication happens via internal `app-network`

## Troubleshooting

### Container Won't Start
```bash
# Check container logs
podman-compose --profile dev logs backend

# Check if ports are in use
ss -tulpn | grep 10801
```

### Qdrant Connection Issues
```bash
# Verify Qdrant is running
podman-compose --profile dev ps

# Test internal connectivity
podman-compose --profile dev exec backend curl http://qdrant:6333/health
```

### Build Issues
```bash
# Clean build cache
podman-compose --profile dev down
podman system prune -f
podman-compose --profile dev up --build
```

## File Structure

```
v_e_x_/                     # Project root
├── podman-compose.yml      # Service definitions
├── .env.dev               # Development environment
├── .env.prod              # Production environment
├── README-podman.md       # This documentation
└── backend/
    ├── Dockerfile         # Multi-stage build
    ├── Cargo.toml         # Rust dependencies
    └── src/               # Source code
```
