#!/bin/bash

# Production deployment script for VEX application
# Uses system environment variables for configuration

set -e  # Exit on any error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

echo "======================================"
echo "VEX Production Deployment"
echo "======================================"
echo ""

# Required environment variables
REQUIRED_VARS=(
    "SERVER_PORT"
    "GIT_USER"
    "GIT_PAT"
    "NOTES_REPO"
    "OPENAI_API_KEY"
)

# Check required environment variables
check_env_vars() {
    local missing_vars=()

    for var in "${REQUIRED_VARS[@]}"; do
        if [[ -z "${!var}" ]]; then
            missing_vars+=("$var")
        fi
    done

    if [[ ${#missing_vars[@]} -gt 0 ]]; then
        print_error "Missing required environment variables:"
        for var in "${missing_vars[@]}"; do
            echo "  - $var"
        done
        echo ""
        echo "Set them with:"
        echo "export SERVER_PORT=22010"
        echo "export GIT_USER=your-username"
        echo "export GIT_PAT=your-token"
        echo "export NOTES_REPO=https://github.com/user/repo"
        echo "export OPENAI_API_KEY=your-key"
        echo ""
        echo "Optional variables:"
        echo "export CLONE_FOLDER=/app/clone"
        echo "export VECTOR_STORAGE_FOLDER=/app/vectors"
        echo "export VOYAGE_API_KEY=your-voyage-key"
        echo "export HARD_CODED_API_KEY=your-api-key"
        return 1
    fi

    # Set defaults for optional variables
    export CLONE_FOLDER="${CLONE_FOLDER:-/app/clone}"
    export VECTOR_STORAGE_FOLDER="${VECTOR_STORAGE_FOLDER:-/app/vectors}"

    return 0
}

# Show configuration (redacted)
show_config() {
    print_status "Production configuration:"
    echo "  SERVER_PORT: ${SERVER_PORT}"
    echo "  GIT_USER: ${GIT_USER}"
    echo "  GIT_PAT: [REDACTED]"
    echo "  CLONE_FOLDER: ${CLONE_FOLDER}"
    echo "  NOTES_REPO: ${NOTES_REPO}"
    echo "  VECTOR_STORAGE_FOLDER: ${VECTOR_STORAGE_FOLDER}"
    echo "  OPENAI_API_KEY: [REDACTED]"
    [[ -n "${VOYAGE_API_KEY}" ]] && echo "  VOYAGE_API_KEY: [REDACTED]"
    [[ -n "${HARD_CODED_API_KEY}" ]] && echo "  HARD_CODED_API_KEY: [REDACTED]"
}

# Health check function
health_check() {
    print_status "Performing health check..."
    local max_attempts=30
    local attempt=1
    local health_url="http://localhost:${SERVER_PORT}/health"

    while [[ $attempt -le $max_attempts ]]; do
        if curl -f -s "$health_url" > /dev/null 2>&1; then
            print_success "Application is healthy and running!"
            return 0
        fi
        echo "Health check attempt $attempt/$max_attempts..."
        sleep 2
        ((attempt++))
    done

    print_error "Health check failed after $max_attempts attempts"
    return 1
}

# Parse command line arguments
case "${1:-deploy}" in
    "deploy"|"up"|"start")
        if ! check_env_vars; then
            exit 1
        fi

        # Detect system architecture for Go build
        print_status "Detecting system architecture..."
        ARCH=$(uname -m)
        case $ARCH in
            x86_64)
                export TARGETARCH=amd64
                ;;
            aarch64|arm64)
                export TARGETARCH=arm64
                ;;
            armv7l)
                export TARGETARCH=arm
                ;;
            *)
                export TARGETARCH=amd64
                print_warning "Unknown architecture $ARCH, defaulting to amd64"
                ;;
        esac
        print_status "System architecture: $ARCH"
        print_status "Target Go architecture: $TARGETARCH"
        echo ""

        show_config
        echo ""

        print_status "Starting production deployment..."

        # Stop existing containers
        print_status "Stopping existing containers..."
        podman-compose -f podman-compose.prod.yml down 2>/dev/null || true

        # Clean up
        print_status "Cleaning up old resources..."
        podman container prune -f || true
        podman image prune -f || true

        # Deploy
        print_status "Building and starting containers..."
        podman-compose -f podman-compose.prod.yml up -d --build

        # Wait for startup
        print_status "Waiting for services to start..."
        sleep 10

        # Verify deployment
        if podman-compose -f podman-compose.prod.yml ps | grep -q "Up"; then
            print_success "Containers started successfully"

            if health_check; then
                print_success "🎉 Production deployment completed successfully!"
                echo ""
                echo "Application available at: http://localhost:${SERVER_PORT}"
                echo ""
                echo "Management commands:"
                echo "  ./prod.sh status - Show container status"
                echo "  ./prod.sh logs   - View application logs"
                echo "  ./prod.sh stop   - Stop the application"
            else
                print_error "Deployment completed but health check failed"
                print_status "Recent logs:"
                podman-compose -f podman-compose.prod.yml logs --tail=20 vex-backend
                exit 1
            fi
        else
            print_error "Failed to start containers"
            podman-compose -f podman-compose.prod.yml ps
            podman-compose -f podman-compose.prod.yml logs vex-backend
            exit 1
        fi
        ;;

    "stop"|"down")
        print_status "Stopping production application..."
        podman-compose -f podman-compose.prod.yml down
        print_success "Production application stopped"
        ;;

    "restart")
        if ! check_env_vars; then
            exit 1
        fi
        print_status "Restarting production application..."
        podman-compose -f podman-compose.prod.yml down
        podman-compose -f podman-compose.prod.yml up -d --build
        sleep 10
        if health_check; then
            print_success "Production application restarted successfully"
        else
            print_error "Restart completed but health check failed"
        fi
        ;;

    "logs")
        print_status "Showing application logs..."
        podman-compose -f podman-compose.prod.yml logs -f vex-backend
        ;;

    "status"|"ps")
        print_status "Container status:"
        podman-compose -f podman-compose.prod.yml ps
        echo ""
        print_status "Recent logs:"
        podman-compose -f podman-compose.prod.yml logs --tail=10 vex-backend
        ;;

    "health")
        if ! check_env_vars; then
            exit 1
        fi
        if health_check; then
            print_success "Application is healthy"
        else
            print_error "Application health check failed"
            exit 1
        fi
        ;;

    "build")
        if ! check_env_vars; then
            exit 1
        fi
        print_status "Building production image..."
        podman-compose -f podman-compose.prod.yml build --no-cache
        print_success "Build completed"
        ;;

    "clean")
        print_status "Cleaning up production environment..."
        podman-compose -f podman-compose.prod.yml down -v
        podman container prune -f
        podman image prune -f
        podman volume prune -f
        print_success "Cleanup completed"
        ;;

    "help"|"--help"|"-h")
        echo "VEX Production Deployment Script"
        echo ""
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  deploy, up   - Deploy the production application (default)"
        echo "  stop, down   - Stop the production application"
        echo "  restart      - Restart the production application"
        echo "  logs         - Show and follow application logs"
        echo "  status, ps   - Show container status and recent logs"
        echo "  health       - Check application health"
        echo "  build        - Build production image"
        echo "  clean        - Clean up containers, images, and volumes"
        echo "  help         - Show this help message"
        echo ""
        echo "Required Environment Variables:"
        for var in "${REQUIRED_VARS[@]}"; do
            echo "  $var"
        done
        echo ""
        echo "Optional Environment Variables:"
        echo "  CLONE_FOLDER (default: /app/clone)"
        echo "  VECTOR_STORAGE_FOLDER (default: /app/vectors)"
        echo "  VOYAGE_API_KEY"
        echo "  HARD_CODED_API_KEY"
        ;;

    *)
        print_error "Unknown command: $1"
        echo "Use '$0 help' to see available commands"
        exit 1
        ;;
esac
