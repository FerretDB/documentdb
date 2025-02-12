#!/bin/bash

set -euo pipefail

# Function to display help message
function show_help {
    echo "Usage: $0 --os <OS> --pg <PG_VERSION> [--test-clean-install] [--output-dir <DIR>] [-h|--help]"
    echo ""
    echo "Description:"
    echo "  This script builds extension packages using Docker."
    echo ""
    echo "Mandatory Arguments:"
    echo "  --os                 OS to build packages for. Possible values: [deb11, deb12, ubuntu20.04, ubuntu22.04, ubuntu24.04]"
    echo "  --pg                 PG version to build packages for. Possible values: [15, 16]"
    echo ""
    echo "Optional Arguments:"
    echo "  --branch             The git branch used for the pre-release built."
    echo "  --hash               The hash of git commit used for the pre-release built."
    echo "  --test-clean-install Test installing the packages in a clean Docker container."
    echo "  --output-dir         Relative path from the repo root of the directory where to drop the packages. The directory will be created if it doesn't exist. Default: packaging"
    echo "  -h, --help           Display this help message."
    exit 0
}

# Initialize variables
OS=""
PG=""
BRANCH=""
HASH=""
TEST_CLEAN_INSTALL=false
OUTPUT_DIR="packaging"  # Default value for output directory

# Process arguments to convert long options to short ones
while [[ $# -gt 0 ]]; do
    case "$1" in
        --os)
            shift
            case $1 in
                deb11|deb12|ubuntu20.04|ubuntu22.04|ubuntu24.04)
                    OS=$1
                    ;;
                *)
                    error_exit "Invalid --os value. Allowed values are [deb11, deb12, ubuntu20.04, ubuntu22.04, ubuntu24.04]"
                    ;;
            esac
            ;;
        --pg)
            shift
            case $1 in
                15|16)
                    PG=$1
                    ;;
                *)
                    error_exit "Invalid --pg value. Allowed values are [15, 16]"
                    ;;
            esac
            ;;
        --branch)
            shift
            BRANCH=$1
            ;;
        --hash)
            shift
            HASH=$1
            ;;
        --test-clean-install)
            TEST_CLEAN_INSTALL=true
            ;;
        --output-dir)
            shift
            OUTPUT_DIR=$1
            ;;
        -h|--help)
            show_help
            ;;
        *)
            echo "Unknown argument: $1"
            show_help
            exit 1
            ;;
    esac
    shift
done

# Check mandatory arguments
if [[ -z "$OS" ]]; then
    echo "Error: --os is required."
    show_help
    exit 1
fi

if [[ -z "$PG" ]]; then
    echo "Error: --pg is required."
    show_help
    exit 1
fi

# Set the appropriate Docker image based on the OS
case $OS in
    deb11)
        DOCKER_IMAGE="debian:bullseye"
        ;;
    deb12)
        DOCKER_IMAGE="debian:bookworm"
        ;;
    ubuntu20.04)
        DOCKER_IMAGE="ubuntu:20.04"
        ;;
    ubuntu22.04)
        DOCKER_IMAGE="ubuntu:22.04"
        ;;
    ubuntu24.04)
        DOCKER_IMAGE="ubuntu:24.04"
        ;;
esac

repo_root=$(git rev-parse --show-toplevel)
abs_output_dir="$repo_root/$OUTPUT_DIR"
cd $repo_root

echo "Building packages for OS: $OS and PostgreSQL version: $PG"
echo "Output directory: $abs_output_dir"

# Create the output directory if it doesn't exist
mkdir -p $abs_output_dir

# Build the Docker image while showing the output to the console
docker build --platform linux/amd64 -t documentdb-build-packages:latest -f ferretdb_packaging/Dockerfile_build_deb_packages \
    --build-arg BASE_IMAGE=$DOCKER_IMAGE --build-arg POSTGRES_VERSION=$PG --build-arg GIT_BRANCH=$BRANCH --build-arg GIT_HASH=$HASH .

# Run the Docker container to build the packages
docker run --platform linux/amd64 --rm -v $abs_output_dir:/output documentdb-build-packages:latest

echo "Packages built successfully!!"

if [[ $TEST_CLEAN_INSTALL == true ]]; then
    echo "Testing clean installation in a Docker container..."

    # Get most recently built debian package for the $PG version
    deb_package_name=$(ls -t $abs_output_dir | grep -E "postgresql-$PG-documentdb.*\.deb" | grep -v "dbg" | head -n 1)
    deb_package_rel_path="$OUTPUT_DIR/$deb_package_name"

    echo "Debian package path: $deb_package_rel_path"

    # Build the Docker image while showing the output to the console
    docker build --platform linux/amd64 -t documentdb-test-packages:latest -f ferretdb_packaging/test_packages/Dockerfile_test_install_deb_packages \
        --build-arg BASE_IMAGE=$DOCKER_IMAGE --build-arg POSTGRES_VERSION=$PG --build-arg DEB_PACKAGE_REL_PATH=$deb_package_rel_path .

    # Run the Docker container to test the packages
    docker run --platform linux/amd64 --rm documentdb-test-packages:latest

    echo "Clean installation test successful!!"
fi

echo "Packages are available in $abs_output_dir"
