# argocd-diff-preview-pr-comment

> **Note**: This project (code and documentation) was created using AI tools (Claude Sonnet 4.5).

[![Release](https://github.com/belitre/argocd-diff-preview-pr-comment/actions/workflows/release.yml/badge.svg)](https://github.com/belitre/argocd-diff-preview-pr-comment/actions/workflows/release.yml)
[![Latest Release](https://img.shields.io/github/v/release/belitre/argocd-diff-preview-pr-comment)](https://github.com/belitre/argocd-diff-preview-pr-comment/releases/latest)
[![License](https://img.shields.io/github/license/belitre/argocd-diff-preview-pr-comment)](LICENSE)

[![Go Version](https://img.shields.io/github/go-mod/go-version/belitre/argocd-diff-preview-pr-comment)](go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/belitre/argocd-diff-preview-pr-comment)](https://goreportcard.com/report/github.com/belitre/argocd-diff-preview-pr-comment)

A Go CLI application that processes ArgoCD application diffs from [argocd-diff-preview](https://github.com/dag-andersen/argocd-diff-preview), splits them by application, and posts organized comments on GitHub pull requests. This tool helps teams review ArgoCD changes more effectively by providing clear, application-specific diff summaries directly in PR comments.

## Features

- 🚀 Process ArgoCD application diffs
- 📦 Split diffs by application for better organization
- 💬 Post structured comments on GitHub PRs
- 🎯 Configurable log levels (debug, info, warn, error, fatal)
- 🔧 Built with Go for performance and reliability
- 📊 Comprehensive test coverage
- 🏗️ Cross-platform support (Linux, macOS)

## Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/belitre/argocd-diff-preview-pr-comment/releases).

Available architectures:
- Linux (amd64, arm64)
- macOS (arm64)

### Build from Source

**Prerequisites:**
- Go 1.25.7 or higher
- Make

```bash
# Clone the repository
git clone https://github.com/belitre/argocd-diff-preview-pr-comment.git
cd argocd-diff-preview-pr-comment

# Build for current platform
make build

# Or build for all supported platforms
make build-cross

# The binaries will be in the ./build directory
```

## Usage

### Post Diffs to GitHub Pull Requests

The `add` command posts split ArgoCD diffs as comments to a GitHub PR:

```bash
# Basic usage with environment variable
export GITHUB_TOKEN=ghp_your_token_here
argocd-diff-preview-pr-comment add \
  --diff-file path/to/diff.txt \
  --pr-ref owner/repo#123

# Using full GitHub PR URL
argocd-diff-preview-pr-comment add \
  --diff-file path/to/diff.txt \
  --pr-ref https://github.com/owner/repo/pull/123 \
  --github-token ghp_your_token_here

# Dry-run to preview without posting
argocd-diff-preview-pr-comment add \
  --diff-file path/to/diff.txt \
  --pr-ref owner/repo#123 \
  --dry-run

# Custom size limit and rate limiting
argocd-diff-preview-pr-comment add \
  --diff-file path/to/diff.txt \
  --pr-ref owner/repo#123 \
  --max-length 32768 \
  --max-retries 5 \
  --retry-delay 3s \
  --backoff-factor 2.5
```

### General Commands

```bash
# Show version and help
argocd-diff-preview-pr-comment --help
argocd-diff-preview-pr-comment version

# Run with custom log level
argocd-diff-preview-pr-comment --log-level debug version

# Available log levels: debug, info, warn, error, fatal
```

### Command-line Flags

#### Add Command (Post to GitHub)

- `--diff-file`: Path to the ArgoCD diff file (required)
- `--pr-ref`: GitHub PR reference in format `owner/repo#123` or full URL (required)
- `--github-token`: GitHub personal access token (optional if using env vars)
- `--max-length`: Maximum length of each comment in bytes (default: 65536)
- `--max-retries`: Maximum number of retry attempts for rate limits (default: 3)
- `--retry-delay`: Initial delay between retries (default: 2s)
- `--backoff-factor`: Exponential backoff multiplier (default: 2.0)
- `--request-timeout`: HTTP request timeout (default: 30s)
- `--dry-run`: Preview actions without posting comments (default: false)
- `--log-level`: Log level (debug, info, warn, error, fatal) (default: "info")

### Rate Limiting

The tool automatically handles GitHub API rate limits:

- **Detection**: Monitors `X-RateLimit-Remaining` header
- **Retry Logic**: Exponential backoff with configurable parameters
- **Default Behavior**: 3 retries with 2s initial delay and 2.0x backoff factor
- **Comment Delay**: 500ms delay between posting comments to avoid rate limits

When rate limited, the tool will:
1. Wait for the time specified in the `X-RateLimit-Reset` header
2. Retry the request with exponential backoff
3. Log detailed information about rate limit status

### CI/CD Integration Example

### GitHub Actions

Use in your GitHub Actions workflow to automatically post ArgoCD diffs on pull requests:

```yaml
name: ArgoCD Diff Preview

on:
  pull_request:
    branches:
      - main

permissions:
  contents: read
  pull-requests: write

jobs:
  argocd-diff:
    name: Generate and Post ArgoCD Diff
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4
        with:
          path: pull-request
      
      - uses: actions/checkout@v4
        with:
          ref: main
          path: main
      
      - name: Prepare secrets for private repos (optional)
        run: |
          mkdir -p secrets
          cat > secrets/secret.yaml << "EOF"
          apiVersion: v1
          kind: Secret
          metadata:
            name: private-repo
            namespace: argocd
            labels:
              argocd.argoproj.io/secret-type: repo-creds
          stringData:
            url: https://github.com/${{ github.repository }}
            password: ${{ secrets.GITHUB_TOKEN }}
            username: not-used
          EOF
      
      - name: Generate Diff
        run: |
          docker run \
            --network=host \
            -v /var/run/docker.sock:/var/run/docker.sock \
            -v $(pwd)/main:/base-branch \
            -v $(pwd)/pull-request:/target-branch \
            -v $(pwd)/output:/output \
            -v $(pwd)/secrets:/secrets \
            -e TARGET_BRANCH=refs/pull/${{ github.event.number }}/merge \
            -e REPO=${{ github.repository }} \
            -e MAX_DIFF_LENGTH=300000 \
            dagandersen/argocd-diff-preview:v0.1.24
      
      - name: Download argocd-diff-preview-pr-comment
        run: |
          LATEST_VERSION=$(curl -s https://api.github.com/repos/belitre/argocd-diff-preview-pr-comment/releases/latest | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
          curl -L -o argocd-diff-preview-pr-comment.tar.gz \
            "https://github.com/belitre/argocd-diff-preview-pr-comment/releases/download/v${LATEST_VERSION}/argocd-diff-preview-pr-comment-${LATEST_VERSION}-linux-amd64.tar.gz"
          tar -xzf argocd-diff-preview-pr-comment.tar.gz
          chmod +x argocd-diff-preview-pr-comment
          sudo mv argocd-diff-preview-pr-comment /usr/local/bin
      
      - name: Post diff as comment
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
            argocd-diff-preview-pr-comment add \
            --file output/diff.md \
            --pr ${{ github.repository }}#${{ github.event.number }} \
            --log-level info
```

## Configuration

The application requires the following environment variables:

- `GITHUB_TOKEN`: GitHub Personal Access Token with `repo` scope (for posting PR comments)
- `GH_TOKEN`: Alternative environment variable for GitHub token (if `GITHUB_TOKEN` is not set)

The GitHub token can also be provided via the `--github-token` flag. If no token is provided through any method, the application will exit with an error.

## Development

### Prerequisites

- Go 1.25.7 or higher
- Make
- GitHub Personal Access Token with `repo` scope

### Available Make Targets

```bash
make help          # Show all available targets
make build         # Build for current platform
make build-cross   # Build for all supported platforms
make test          # Run tests with verbose output
make fmt           # Format code
make lint          # Run linter
make coverage      # Generate coverage report
make clean         # Remove build artifacts
make tidy          # Run go mod tidy
make all           # Run fmt, lint, test, and build-cross
```

### Development Workflow

1. Make code changes following the [code style guidelines](docs/CLAUDE.md#code-style)
2. Run `make fmt` to format code
3. Run `make lint` to check for issues
4. Run `make test` to verify tests pass
5. Run `make coverage` to ensure adequate test coverage
6. Run `make build` to verify compilation
7. Commit changes with clear, descriptive messages following [Conventional Commits](https://www.conventionalcommits.org/)

### Running Tests

```bash
# Run all tests
make test

# Generate coverage report
make coverage

# View coverage in browser
open build/coverage.html
```

## CI/CD Workflows

### CI Workflow (`.github/workflows/ci.yml`)

Triggered on:
- Pull requests to `main` or `develop`
- Pushes to `main` or `develop`

Actions:
1. **Test and Lint Job**:
   - Runs `make tidy` and checks for uncommitted changes
   - Runs `make fmt` and verifies formatting
   - Runs `make lint` to check code quality
   - Runs `make test` to execute all tests
   - Runs `make coverage` to generate coverage reports
   - Runs `make build` to verify local compilation
   - Uploads coverage reports as artifacts

2. **Cross-platform Build Job**:
   - Runs `make build-cross` to build for all platforms
   - Uploads all binaries as artifacts

Both jobs cache Go dependencies for faster builds.

### Release Workflow (`.github/workflows/release.yml`)

Triggered on:
- Pushes to `main` branch

Actions:
1. Runs `make test` and `make lint`
2. Uses [semantic-release](https://semantic-release.gitbook.io/) to:
   - Analyze commit messages (follows [Conventional Commits](https://www.conventionalcommits.org/))
   - Determine the next version number
   - Generate release notes
   - Create a GitHub release
   - Update CHANGELOG.md
3. Builds binaries for all platforms using `make build-cross`
4. Packages binaries as `.tar.gz` and `.zip` files
5. Uploads all archives to the GitHub release

**Commit Message Format:**
- `feat:` → Minor version bump (new features)
- `fix:` → Patch version bump (bug fixes)
- `BREAKING CHANGE:` → Major version bump
- `major:` → Major version bump
- Other types: `docs:`, `refactor:`, `perf:`, `test:`, `chore:`, `ci:`

**Example Commits:**
```bash
git commit -m "feat: add support for multiple repositories"
git commit -m "fix: handle empty diff responses correctly"
git commit -m "docs: update installation instructions"
```

### Caching Strategy

Both workflows cache:
- Go modules (`~/go/pkg/mod`)
- Go build cache (`~/.cache/go-build`)

Cache key is based on `go.sum` content, ensuring dependencies are only re-downloaded when they change.

## Project Structure

```
argocd-diff-preview-pr-comment/
├── cmd/
│   └── argocd-diff-preview-pr-comment/  # Main application entry point
│       ├── main.go                       # CLI commands using Cobra
│       └── main_test.go                  # CLI tests
├── pkg/
│   ├── logger/                           # Logging package (zap)
│   │   ├── logger.go
│   │   └── logger_test.go
│   └── version/                          # Version information
│       ├── version.go
│       └── version_test.go
├── .github/
│   └── workflows/
│       ├── ci.yml                        # CI workflow
│       └── release.yml                   # Release workflow
├── build/                                # Build output directory
├── docs/                                 # Documentation
│   ├── CLAUDE.md                         # Developer guidelines
│   └── IMPLEMENTATION.md                 # GitHub integration details
├── Makefile                              # Build automation
├── go.mod                                # Go module definition
├── go.sum                                # Go module checksums
├── .releaserc.json                       # Semantic-release config
├── CHANGELOG.md                          # Auto-generated changelog
└── README.md                             # This file
```

## Documentation

- [Developer Guidelines](docs/CLAUDE.md) - Code style, testing requirements, and development workflow
- [Implementation Details](docs/IMPLEMENTATION.md) - GitHub PR comment integration implementation
- [Test Coverage Summary](docs/TEST_COVERAGE_SUMMARY.md) - Test coverage status and quality improvements

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/amazing-feature`)
3. Make your changes following the [code style guidelines](docs/CLAUDE.md#code-style)
4. Ensure all tests pass (`make test`)
5. Commit your changes using [Conventional Commits](https://www.conventionalcommits.org/)
6. Push to your branch (`git push origin feat/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the terms specified in the [LICENSE](LICENSE) file.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI functionality
- Uses [zap](https://github.com/uber-go/zap) for high-performance logging
- Integrates with [argocd-diff-preview](https://github.com/dag-andersen/argocd-diff-preview)
- Automated releases with [semantic-release](https://semantic-release.gitbook.io/)