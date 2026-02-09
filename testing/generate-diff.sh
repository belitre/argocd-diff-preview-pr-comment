#!/usr/bin/env bash

set -euo pipefail

# Script to download argocd-diff-preview and generate diffs for testing
# Supports Linux and macOS

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY_NAME="argocd-diff-preview"
GITHUB_REPO="dag-andersen/argocd-diff-preview"
VERSION="latest"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    local os=""
    local arch=""

    # Detect OS (GitHub releases use capitalized names)
    case "$(uname -s)" in
        Linux*)     os="Linux" ;;
        Darwin*)    os="Darwin" ;;
        *)
            log_error "Unsupported operating system: $(uname -s)"
            log_error "This script only supports Linux and macOS"
            exit 1
            ;;
    esac

    # Detect architecture (GitHub releases use specific naming)
    case "$(uname -m)" in
        x86_64)     arch="x86_64" ;;
        aarch64)    arch="aarch64" ;;
        arm64)      arch="aarch64" ;;
        *)
            log_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac

    echo "${os}-${arch}"
}

# Download argocd-diff-preview binary
download_binary() {
    local platform="$1"
    local binary_path="${SCRIPT_DIR}/${BINARY_NAME}"

    # Check if binary already exists
    if [ -f "$binary_path" ]; then
        log_info "Binary already exists at $binary_path"
        log_info "To re-download, delete the existing binary first"
        return 0
    fi

    log_info "Detecting latest release version..."

    # Get latest release info from GitHub API
    local latest_release_url="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
    local download_url=""

    # Get the download URL for the detected platform
    # Format: argocd-diff-preview-Darwin-aarch64.tar.gz
    download_url=$(curl -s "$latest_release_url" | \
        grep "browser_download_url.*${BINARY_NAME}-${platform}.tar.gz" | \
        cut -d '"' -f 4 | \
        head -n 1)

    if [ -z "$download_url" ]; then
        log_error "Could not find download URL for platform: $platform"
        log_error "Please check https://github.com/${GITHUB_REPO}/releases"
        exit 1
    fi

    log_info "Downloading argocd-diff-preview from: $download_url"

    # Download the tar.gz file
    local tar_file="${SCRIPT_DIR}/${BINARY_NAME}.tar.gz"
    if ! curl -L -o "$tar_file" "$download_url"; then
        log_error "Failed to download binary"
        exit 1
    fi

    # Extract the binary
    log_info "Extracting binary..."
    if ! tar -xzf "$tar_file" -C "$SCRIPT_DIR"; then
        log_error "Failed to extract binary"
        rm -f "$tar_file"
        exit 1
    fi

    # Clean up tar file
    rm -f "$tar_file"

    # Make it executable
    chmod +x "$binary_path"

    log_info "Successfully downloaded argocd-diff-preview to $binary_path"
}

# Initialize git repository if needed
init_git_repo() {
    cd "$SCRIPT_DIR" || exit 1

    if [ ! -d ".git" ]; then
        log_info "Initializing git repository..."
        git init
        git add .
        git commit -m "Initial commit with ArgoCD application"
        log_info "Git repository initialized"
    else
        log_info "Git repository already exists"
    fi
}

# Generate diff using argocd-diff-preview
generate_diff() {
    local binary_path="${SCRIPT_DIR}/${BINARY_NAME}"
    local repo_owner="${REPO_OWNER:-your-org}"
    local repo_name="${REPO_NAME:-your-repo}"

    # Get current branch as target branch
    local target_branch
    target_branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "main")

    # Allow override via TARGET_BRANCH env var if needed
    target_branch="${TARGET_BRANCH:-$target_branch}"

    local base_branch="${BASE_BRANCH:-main}"

    if [ ! -f "$binary_path" ]; then
        log_error "Binary not found at $binary_path"
        exit 1
    fi

    log_info "Repository: ${repo_owner}/${repo_name}"
    log_info "Target branch: $target_branch (current branch)"
    log_info "Base branch: $base_branch"
    log_info "Max diff length: 999999"
    log_info ""
    log_warn "Note: argocd-diff-preview requires a GitHub repository with ArgoCD applications."
    log_warn "Set environment variables: REPO_OWNER, REPO_NAME, BASE_BRANCH (optional)"
    log_info ""

    # Run argocd-diff-preview
    # This tool works with GitHub repositories, not local files
    if ! "$binary_path" \
        --repo "${repo_owner}/${repo_name}" \
        --target-branch "$target_branch" \
        --base-branch "$base_branch" \
        --max-diff-length 999999; then
        log_error "Failed to generate diff"
        exit 1
    fi

    log_info ""
    log_info "Diff generation completed successfully"
}

# Main execution
main() {
    log_info "ArgoCD Diff Preview - Test Diff Generator"
    log_info "=========================================="

    # Detect platform
    local platform
    platform=$(detect_platform)
    log_info "Detected platform: $platform"

    # Download binary if needed
    download_binary "$platform"

    # Initialize git repository if needed (for local testing)
    init_git_repo

    # Generate diff
    generate_diff

    log_info "Done!"
}

main "$@"
