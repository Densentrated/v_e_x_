# V_E_X: A Chatbot that utilizes YOUR vectorized obsidian notes, remembers what you don't

V_E_X is a Go-based backend service that creates a chatbot using your Obsidian notes. It vectorizes your notes and provides intelligent responses based on your knowledge base.

## Features

- 🔍 **Vector Search**: Searches through your vectorized Obsidian notes
- 🤖 **AI Chat**: Integrates with OpenAI for intelligent responses
- 📚 **Knowledge Base**: Uses your personal notes as the knowledge source
- 🔒 **Secure**: API key authentication and secure deployment
- 🐳 **Containerized**: Easy deployment with Podman/Docker

## Prerequisites

- Go 1.21+
- Podman or Docker
- OpenAI API key
- Git repository with your Obsidian notes

## Quick Start

### Development (Local)

1. **Clone the repository**
   ```bash
   git clone <your-repo-url>
   cd v_e_x_
   ```

2. **Set up environment**
   ```bash
   # Copy example environment file
   cp "example .env" .env
   
   # Edit .env with your configuration
   nano .env
   ```

3. **Start development environment**
   ```bash
   ./dev.sh
   ```

4. **Access the application**
   - API: http://localhost:22010
   - Health check: http://localhost:22010/health

### Production Deployment

1. **Set environment variables**
   ```bash
   export SERVER_PORT=22010
   export GIT_USER=your-git-username
   export GIT_PAT=your-personal-access-token
   export NOTES_REPO=https://github.com/user/your-notes-repo
   export OPENAI_API_KEY=your-openai-api-key
   
   # Optional variables
   export CLONE_FOLDER=/app/clone
   export VECTOR_STORAGE_FOLDER=/app/vectors
   export VOYAGE_API_KEY=your-voyage-api-key
   export HARD_CODED_API_KEY=your-api-key
   ```

2. **Deploy**
   ```bash
   ./prod.sh deploy
   ```

## Configuration

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `SERVER_PORT` | Port for the server | `22010` |
| `GIT_USER` | Git username | `your-username` |
| `GIT_PAT` | Personal access token | `ghp_...` |
| `NOTES_REPO` | Your notes repository URL | `https://github.com/user/notes` |
| `OPENAI_API_KEY` | OpenAI API key | `sk-...` |

### Optional Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CLONE_FOLDER` | Local clone directory | `/app/clone` |
| `VECTOR_STORAGE_FOLDER` | Vector storage directory | `/app/vectors` |
| `VOYAGE_API_KEY` | Voyage AI API key | - |
| `HARD_CODED_API_KEY` | API key for authentication | - |

## Development Scripts

### `./dev.sh` - Development Environment

Uses `.env` file for configuration and `podman-compose.dev.yml`.

```bash
# Start development environment
./dev.sh

# View logs
./dev.sh logs

# Stop environment
./dev.sh stop

# Restart environment
./dev.sh restart

# Access container shell
./dev.sh shell

# Show help
./dev.sh help
```

### `./prod.sh` - Production Deployment

Uses system environment variables and `podman-compose.prod.yml`.

```bash
# Deploy to production
./prod.sh deploy

# Check status
./prod.sh status

# View logs
./prod.sh logs

# Health check
./prod.sh health

# Stop production
./prod.sh stop

# Show help
./prod.sh help
```

## Compose Files

- **`podman-compose.yml`**: Main production configuration
- **`podman-compose.prod.yml`**: Production-specific configuration (uses system env vars)
- **`podman-compose.dev.yml`**: Development configuration (uses `.env` file)

## API Endpoints

### Health Check
```bash
GET /health
```

Returns application health status.

### Chat Endpoint
```bash
POST /chat
Content-Type: application/json
Authorization: Bearer <your-api-key>

{
  "message": "Your question here"
}
```

## CI/CD Deployment

The project includes automated deployment via Gitea Actions in `.gitea/workflows/deploy.yml`.

### Required Secrets

Set these secrets in your Gitea repository:

| Secret | Description |
|--------|-------------|
| `SSH_PRIVATE_KEY` | SSH private key for VPS access |
| `SSH_USER` | SSH username for VPS |
| `VPS_HOST` | VPS hostname or IP |
| `SERVER_PORT` | Application port (default: 22010) |
| `GIT_USER` | Git username |
| `GIT_PAT` | Git personal access token |
| `NOTES_REPO` | Your notes repository URL |
| `OPENAI_API_KEY` | OpenAI API key |
| `VOYAGE_API_KEY` | (Optional) Voyage API key |
| `HARD_CODED_API_KEY` | (Optional) API key for authentication |

### Deployment Process

1. **Push to main branch** triggers deployment
2. **Files are copied** to VPS
3. **Environment variables** are set from secrets
4. **Application builds** and deploys using `podman-compose.prod.yml`
5. **Health checks** verify successful deployment

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         V_E_X                              │
├─────────────────────────────────────────────────────────────┤
│ Frontend (Optional)                                         │
├─────────────────────────────────────────────────────────────┤
│ Go Backend                                                  │
│  ├── Chat Handler                                          │
│  ├── Vector Search                                         │
│  ├── Git Integration                                       │
│  └── OpenAI Integration                                    │
├─────────────────────────────────────────────────────────────┤
│ Data Layer                                                  │
│  ├── Vectorized Notes (LanceDB)                           │
│  ├── Git Repository (Notes)                               │
│  └── Vector Storage                                        │
└─────────────────────────────────────────────────────────────┘
```

## Project Structure

```
v_e_x_/
├── backend/
│   ├── chat/          # Chat handling logic
│   ├── config/        # Configuration management
│   ├── git/           # Git operations
│   ├── handlers/      # HTTP handlers
│   ├── routes/        # API routes
│   ├── vector/        # Vector operations
│   └── main.go        # Application entry point
├── .gitea/workflows/  # CI/CD workflows
├── Dockerfile         # Container image definition
├── podman-compose*.yml # Container orchestration
├── dev.sh            # Development script
├── prod.sh           # Production script
└── example .env      # Environment template
```

## Troubleshooting

### Common Issues

1. **Missing .env file in development**
   ```bash
   # Copy from example
   cp "example .env" .env
   # Edit with your values
   nano .env
   ```

2. **Container build failures**
   ```bash
   # Clean up and rebuild
   ./dev.sh clean
   ./dev.sh build
   ```

3. **Health check failures**
   ```bash
   # Check logs
   ./prod.sh logs
   # Check status
   ./prod.sh status
   ```

4. **Environment variable issues**
   ```bash
   # Verify all required variables are set
   ./prod.sh help
   ```

### Logs and Monitoring

```bash
# Development logs
./dev.sh logs

# Production logs
./prod.sh logs

# Container status
podman ps
podman-compose ps

# System resources
podman stats
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test with `./dev.sh`
5. Submit a pull request

## License

See LICENSE file for details.

## Support

- Create an issue for bugs or feature requests
- Check logs with `./dev.sh logs` or `./prod.sh logs`
- Verify configuration with environment variable documentation above