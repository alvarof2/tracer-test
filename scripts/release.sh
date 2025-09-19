#!/bin/bash

# Release script for tracer-test
# Usage: ./scripts/release.sh <version>
# Example: ./scripts/release.sh v1.0.0

set -e

# Colors for output
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

# Check if version is provided
if [ $# -eq 0 ]; then
    print_error "Version is required!"
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

VERSION=$1

# Validate version format
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    print_error "Invalid version format. Use semantic versioning (e.g., v1.0.0)"
    exit 1
fi

print_status "Starting release process for version: $VERSION"

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    print_error "Not in a git repository!"
    exit 1
fi

# Check if working directory is clean
if ! git diff-index --quiet HEAD --; then
    print_error "Working directory is not clean. Please commit or stash your changes."
    exit 1
fi

# Check if tag already exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    print_error "Tag $VERSION already exists!"
    exit 1
fi

# Run tests
print_status "Running tests..."
if ! make test; then
    print_error "Tests failed!"
    exit 1
fi
print_success "Tests passed!"

# Run linter
print_status "Running linter..."
if ! make lint; then
    print_warning "Linter found issues, but continuing..."
fi

# Build for all platforms
print_status "Building for all platforms..."
if ! make build-all; then
    print_error "Build failed!"
    exit 1
fi
print_success "Build completed!"

# Create checksums
print_status "Creating checksums..."
make checksums
print_success "Checksums created!"

# Create git tag
print_status "Creating git tag: $VERSION"
git tag -a "$VERSION" -m "Release $VERSION"
print_success "Git tag created!"

# Push tag to remote
print_status "Pushing tag to remote..."
if ! git push origin "$VERSION"; then
    print_error "Failed to push tag to remote!"
    exit 1
fi
print_success "Tag pushed to remote!"

# Show release summary
print_success "Release $VERSION created successfully!"
echo ""
echo "Release Summary:"
echo "  Version: $VERSION"
echo "  Tag: $VERSION"
echo "  Binaries: dist/"
echo "  Checksums: dist/*.sha256"
echo ""
echo "Next steps:"
echo "  1. GitHub Actions will automatically create a release"
echo "  2. Check the Actions tab in your GitHub repository"
echo "  3. The release will include all platform binaries and checksums"
echo ""
echo "To create a release manually:"
echo "  gh release create $VERSION dist/* --title \"Release $VERSION\" --notes \"Release $VERSION\""
echo ""
echo "To delete this release (if needed):"
echo "  git tag -d $VERSION"
echo "  git push origin :refs/tags/$VERSION"
