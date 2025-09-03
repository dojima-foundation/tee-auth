#!/bin/bash

# Environment detection script for Renclave testing
# This script provides functions to detect the testing environment

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to detect operating system
detect_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux";;
        Darwin*)    echo "macos";;
        CYGWIN*)   echo "windows";;
        MINGW*)    echo "windows";;
        *)          echo "unknown";;
    esac
}

# Function to get test capabilities
get_test_capabilities() {
    local os=$(detect_os)

    case "$os" in
        "linux")
            echo "full"  # Can run all tests natively
            ;;
        "macos")
            # macOS can run unit tests locally, but needs Docker for enclave/host
            if check_docker && check_docker_compose; then
                echo "hybrid"
            else
                echo "limited" # Only local unit tests
            fi
            ;;
        "windows")
            # Windows (WSL2) can run unit tests locally, needs Docker for enclave/host
            if check_docker && check_docker_compose; then
                echo "hybrid"
            else
                echo "limited" # Only local unit tests
            fi
            ;;
        *)
            echo "limited" # Unknown OS, assume limited capabilities
            ;;
    esac
}

# Function to check if Docker is running
check_docker() {
    docker info > /dev/null 2>&1
    return $?
}

# Function to check if Docker Compose is available
check_docker_compose() {
    docker compose version > /dev/null 2>&1
    return $?
}

# Function to determine the test strategy
get_test_strategy() {
    local unit_only_flag=$1
    local capabilities=$(get_test_capabilities)

    if [ "$unit_only_flag" = "--unit-only" ]; then
        echo "local-unit"
    elif [ "$capabilities" = "full" ]; then
        echo "full-local"
    elif [ "$capabilities" = "hybrid" ]; then
        echo "hybrid-docker"
    else
        echo "local-unit" # Fallback to local unit tests
    fi
}

# Function to print status
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to print environment summary
print_environment_summary() {
    local os=$(detect_os)
    local capabilities=$(get_test_capabilities)
    local docker_status="Not Running"
    local docker_compose_status="Not Found"

    if check_docker; then
        docker_status="Running"
    fi
    if check_docker_compose; then
        docker_compose_status="Found (v2)"
    fi

    print_status $BLUE "Environment Summary:"
    print_status $BLUE "  OS: $os"
    print_status $BLUE "  Docker: $docker_status"
    print_status $BLUE "  Docker Compose: $docker_compose_status"
    print_status $BLUE "  Test Capabilities: $capabilities"
    echo ""
}
