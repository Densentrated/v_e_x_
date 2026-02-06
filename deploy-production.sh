#!/bin/bash

# Production deployment script for VEX application
# This script sets up environment variables and deploys the application

set -e  # Exit on any error

echo "🚀 Starting VEX production deployment..."

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

# Check if required environment variables are set
check_env_vars() {
    local missing_vars=()

    # Required environment variables
    local required_vars=(
        "SERVER_PORT"
        "GIT_USER"
        "GIT_PAT"
        "NOTES_REPO"
        "OPENAI_API_KEY"
    )

    for var in "${required_vars[@]}"; do
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
        echo "Please set these variables before running the deployment:"
        echo "export SERVER_PORT=22010"
        echo "export GIT_USER=your-git-username"
        echo "export GIT_PAT=your-personal-access-token"
        echo "export NOTES_REPO=https://github.com/user/repo"
        echo "export OPENAI_API_KEY=your-openai-api-key"
        echo ""
        echo "Optional variables:"
        echo "export CLONE_FOLDER=/app/clone"
        echo "export VECTOR_STORAGE_FOLDER=/app/vectors"
        echo "export VOYAGE_API_KEY=your-voyage-api-key"
        echo "export HARD_CODED_API_KEY=your-api-key"
        exit 1
    fi
}

# Set default values for optional environment variables
set_defaults() {
    export CLONE_FOLDER="${CLONE_FOLDER:-/app/clone}"
    export VECTOR_STORAGE_FOLDER="${VECTOR_STORAGE_FOLDER:-/app/vectors}"

    print_status "Environment variables configured:"
    echo "  SERVER_PORT: ${SERVER_PORT}"
    echo "  GIT_USER: ${GIT_USER}"
    echo "  GIT_PAT: [REDACTED]"
    echo "  CLONE_FOLDER: ${CLONE_FOLDER}"
    echo "  NOTES_REPO: ${NOTES_REPO}"
    echo "  VECTOR_STORAGE_FOLDER: ${VECTOR_STORAGE_FOLDER}"
    echo "  OPENAI_API_KEY: [REDACTED]"
    if [[ -n "${VOYAGE_API_KEY}" ]]; then
        echo "  VOYAGE_API_KEY: [REDACTED]"
    fi
    if [[ -n "${HARD_CODED_API_KEY}" ]]; then
        echo "  HARD_CODED_API_KEY: [REDACTED]"
    fi
}

# Stop existing containers
stop_containers() {
    print_status "Stopping existing containers..."

    # Stop and remove containers gracefully
    if podman-compose ps | grep -q "vex-backend"; then
        podman-compose down
        print_success "Stopped existing containers"
    else
        print_warning "No existing containers found"
    fi
}

# Clean up old images (optional)
cleanup_images() {
    print_status "Cleaning up old images..."

    # Remove dangling images
    dangling_images=$(podman images -f "dangling=true" -q)
    if [[ -n "$dangling_images" ]]; then
        podman rmi $dangling_images || true
        print_success "Removed dangling images"
    else
        print_status "No dangling images to remove"
    fi
}

# Build and deploy
deploy() {
    print_status "Building and deploying VEX application..."

    # Build and start containers
    podman-compose up -d --build

    if [[ $? -eq 0 ]]; then
        print_success "Deployment completed successfully!"
    else
        print_error "Deployment failed!"
        exit 1
    fi
}

# Health check
health_check() {
    print_status "Performing health check..."

    local max_attempts=30
    local attempt=1
    local health_url="http://localhost:${SERVER_PORT}/health"

    while [[ $attempt -le $max_attempts ]]; do
        print_status "Health check attempt $attempt/$max_attempts..."

        if curl -f -s "$health_url" > /dev/null 2>&1; then
            print_success "Application is healthy and running!"
            print_success "Application available at: http://localhost:${SERVER_PORT}"
            return 0
        fi

        sleep 2
        ((attempt++))
    done

    print_error "Health check failed after $max_attempts attempts"
    print_warning "Check container logs with: podman-compose logs vex-backend"
    return 1
}

# Show container status
show_status() {
    print_status "Container status:"
    podman-compose ps
    echo ""

    print_status "Recent logs:"
    podman-compose logs --tail=20 vex-backend
}

# Main deployment flow
main() {
    echo "======================================"
    echo "VEX Production Deployment Script"
    echo "======================================"
    echo ""

    # Check environment variables
    check_env_vars

    # Set defaults
    set_defaults

    echo ""
    read -p "Continue with deployment? (y/N): " -n 1 -r
    echo ""

    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Deployment cancelled by user"
        exit 0
    fi

    # Execute deployment steps
    stop_containers
    cleanup_images
    deploy

    # Wait a moment for containers to start
    sleep 5

    # Health check
    if health_check; then
        show_status
        print_success "🎉 VEX application deployed successfully!"
        echo ""
        echo "Next steps:"
        echo "- Monitor logs: podman-compose logs -f vex-backend"
        echo "- Check status: podman-compose ps"
        echo "- Stop service: podman-compose down"
    else
        print_error "Deployment completed but health check failed"
        show_status
        exit 1
    fi
}

# Handle script interruption
trap 'print_error "Deployment interrupted"; exit 1' INT TERM

# Run main function
main "$@"
