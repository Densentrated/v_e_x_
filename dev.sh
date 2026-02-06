#!/bin/bash

# Development script for VEX application
# Uses .env file for configuration

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
echo "VEX Development Environment"
echo "======================================"
echo ""

# Check if .env file exists
if [ ! -f .env ]; then
    print_warning ".env file not found!"
    echo ""
    echo "Creating .env file from example..."

    if [ -f "example .env" ]; then
        cp "example .env" .env
        print_success "Created .env file from example"
    elif [ -f ".env.example" ]; then
        cp ".env.example" .env
        print_success "Created .env file from .env.example"
        print_warning "Please edit .env file with your configuration before continuing"
        echo ""
        echo "Required variables to configure:"
        echo "  - GIT_USER: Your git username"
        echo "  - GIT_PAT: Your personal access token"
        echo "  - NOTES_REPO: Your notes repository URL"
        echo "  - OPENAI_API_KEY: Your OpenAI API key"
        echo ""
        echo "Edit .env file now? (y/N):"
        read -r edit_env
        if [[ $edit_env =~ ^[Yy]$ ]]; then
            ${EDITOR:-nano} .env
        fi
    else
        print_error "Neither 'example .env' nor '.env.example' file found!"
        print_error "Please create a .env file manually with required configuration"
        exit 1
    fi
fi

print_status "Using .env file for configuration"

# Show current configuration (redacted)
echo ""
print_status "Current configuration:"
while IFS='=' read -r key value; do
    # Skip empty lines and comments
    [[ -z "$key" || "$key" =~ ^#.*$ ]] && continue

    # Redact sensitive values
    if [[ "$key" =~ (PAT|KEY|TOKEN) ]]; then
        echo "  $key: [REDACTED]"
    else
        echo "  $key: $value"
    fi
done < .env

echo ""

# Parse command line arguments
case "${1:-up}" in
    "up"|"start")
        print_status "Starting development environment..."
        podman-compose -f podman-compose.dev.yml up -d --build

        # Wait for services to start
        sleep 5

        # Check if services are running
        if podman-compose -f podman-compose.dev.yml ps | grep -q "Up"; then
            print_success "Development environment started successfully!"
            echo ""
            echo "🚀 Application available at: http://localhost:$(grep SERVER_PORT .env | cut -d'=' -f2 | tr -d ' ')"
            echo ""
            echo "Useful commands:"
            echo "  ./dev.sh logs    - View application logs"
            echo "  ./dev.sh stop    - Stop the application"
            echo "  ./dev.sh restart - Restart the application"
            echo "  ./dev.sh shell   - Access container shell"
        else
            print_error "Failed to start development environment"
            print_status "Container status:"
            podman-compose -f podman-compose.dev.yml ps
            print_status "Checking logs..."
            podman-compose -f podman-compose.dev.yml logs vex-backend
            exit 1
        fi
        ;;

    "down"|"stop")
        print_status "Stopping development environment..."
        podman-compose -f podman-compose.dev.yml down
        print_success "Development environment stopped"
        ;;

    "restart")
        print_status "Restarting development environment..."
        podman-compose -f podman-compose.dev.yml down
        podman-compose -f podman-compose.dev.yml up -d --build
        sleep 5
        print_success "Development environment restarted"
        ;;

    "logs")
        print_status "Showing application logs..."
        podman-compose -f podman-compose.dev.yml logs -f vex-backend
        ;;

    "shell"|"exec")
        print_status "Accessing container shell..."
        podman-compose -f podman-compose.dev.yml exec vex-backend sh
        ;;

    "ps"|"status")
        print_status "Container status:"
        podman-compose -f podman-compose.dev.yml ps
        ;;

    "build")
        print_status "Building application..."
        podman-compose -f podman-compose.dev.yml build --no-cache
        print_success "Build completed"
        ;;

    "clean")
        print_status "Cleaning up development environment..."
        podman-compose -f podman-compose.dev.yml down -v
        podman container prune -f
        podman image prune -f
        print_success "Cleanup completed"
        ;;

    "help"|"--help"|"-h")
        echo "VEX Development Script"
        echo ""
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  up, start    - Start the development environment (default)"
        echo "  down, stop   - Stop the development environment"
        echo "  restart      - Restart the development environment"
        echo "  logs         - Show and follow application logs"
        echo "  shell, exec  - Access the container shell"
        echo "  ps, status   - Show container status"
        echo "  build        - Rebuild the application"
        echo "  clean        - Clean up containers and images"
        echo "  help         - Show this help message"
        echo ""
        echo "Examples:"
        echo "  $0           - Start development environment"
        echo "  $0 logs      - View logs"
        echo "  $0 restart   - Restart services"
        ;;

    *)
        print_error "Unknown command: $1"
        echo "Use '$0 help' to see available commands"
        exit 1
        ;;
esac
