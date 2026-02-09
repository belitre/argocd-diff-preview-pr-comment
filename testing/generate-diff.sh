#!/usr/bin/env bash

set -euo pipefail
set -x  # Enable command tracing

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

# Prepare GitHub token secret
prepare_secret() {
    local repo_owner="$1"
    local repo_name="$2"
    local secrets_dir="${SCRIPT_DIR}/secrets"
    local secret_file="${secrets_dir}/github-token.yaml"

    # Check if secret already exists
    if [ -f "$secret_file" ]; then
        log_info "GitHub token secret already exists at $secret_file"
        return 0
    fi

    # Secret doesn't exist, check for GITHUB_TOKEN env var
    if [ -z "${GITHUB_TOKEN:-}" ]; then
        log_error "GitHub token secret not found and GITHUB_TOKEN environment variable is not set"
        log_error "Please set GITHUB_TOKEN environment variable or create the secret manually at:"
        log_error "  $secret_file"
        exit 1
    fi

    # Create secrets directory if it doesn't exist
    mkdir -p "$secrets_dir"

    log_info "Creating GitHub token secret..."

    # Create the secret file
    cat > "$secret_file" <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: private-repo
  namespace: argocd
  labels:
    argocd.argoproj.io/secret-type: repo-creds
stringData:
  url: https://github.com/${repo_owner}/${repo_name}
  password: ${GITHUB_TOKEN}
  username: not-used
EOF

    log_info "GitHub token secret created successfully at $secret_file"
}

# Prepare branch directories
prepare_branches() {
    local repo_url="$1"
    local base_branch="$2"
    local target_branch="$3"

    local base_dir="${SCRIPT_DIR}/base-branch"
    local target_dir="${SCRIPT_DIR}/target-branch"

    # Prepare base branch directory
    log_info "Preparing base branch directory: $base_branch"
    if [ -d "$base_dir" ]; then
        log_info "Cleaning existing base-branch directory..."
        rm -rf "$base_dir"
    fi

    log_info "Cloning base branch ($base_branch)..."
    if ! git clone --branch "$base_branch" --depth 1 "$repo_url" "$base_dir"; then
        log_error "Failed to clone base branch"
        exit 1
    fi

    # Prepare target branch directory
    log_info "Preparing target branch directory: $target_branch"
    if [ -d "$target_dir" ]; then
        log_info "Cleaning existing target-branch directory..."
        rm -rf "$target_dir"
    fi

    log_info "Cloning target branch ($target_branch)..."
    if ! git clone --branch "$target_branch" --depth 1 "$repo_url" "$target_dir"; then
        log_error "Failed to clone target branch"
        exit 1
    fi

    log_info "Branch directories prepared successfully"
}

# Generate diff using argocd-diff-preview
generate_diff() {
    local binary_path="${SCRIPT_DIR}/${BINARY_NAME}"
    local repo_owner="${REPO_OWNER:-belitre}"
    local repo_name="${REPO_NAME:-argocd-diff-preview-pr-comment}"
    local repo_url="https://github.com/${repo_owner}/${repo_name}.git"

    # Get current branch from the parent repository (not the testing folder)
    local target_branch
    target_branch=$(cd "${SCRIPT_DIR}/.." && git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "main")

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

    # Prepare GitHub token secret
    prepare_secret "$repo_owner" "$repo_name"

    log_info ""

    # Prepare local branch directories
    prepare_branches "$repo_url" "$base_branch" "$target_branch"

    log_info ""
    log_info "Running argocd-diff-preview..."
    log_info ""

    # Run argocd-diff-preview with local directories
    if ! "$binary_path" \
        --repo "${repo_owner}/${repo_name}" \
        --base-branch "${SCRIPT_DIR}/base-branch" \
        --target-branch "${SCRIPT_DIR}/target-branch" \
        --max-diff-length 400 --debug --argocd-chart-version 9.1.9; then
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

    # Generate diff
    generate_diff

    log_info "Done!"
}

main "$@"
