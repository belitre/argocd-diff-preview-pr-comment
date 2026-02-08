# argocd-diff-preview-pr-comment

[![CI](https://github.com/belitre/argocd-diff-preview-pr-comment/actions/workflows/ci.yml/badge.svg)](https://github.com/belitre/argocd-diff-preview-pr-comment/actions/workflows/ci.yml)
[![Release](https://github.com/belitre/argocd-diff-preview-pr-comment/actions/workflows/release.yml/badge.svg)](https://github.com/belitre/argocd-diff-preview-pr-comment/actions/workflows/release.yml)
[![Latest Release](https://img.shields.io/github/v/release/belitre/argocd-diff-preview-pr-comment)](https://github.com/belitre/argocd-diff-preview-pr-comment/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/belitre/argocd-diff-preview-pr-comment)](go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/belitre/argocd-diff-preview-pr-comment)](https://goreportcard.com/report/github.com/belitre/argocd-diff-preview-pr-comment)
[![License](https://img.shields.io/github/license/belitre/argocd-diff-preview-pr-comment)](LICENSE)
[![GitHub issues](https://img.shields.io/github/issues/belitre/argocd-diff-preview-pr-comment)](https://github.com/belitre/argocd-diff-preview-pr-comment/issues)
[![GitHub stars](https://img.shields.io/github/stars/belitre/argocd-diff-preview-pr-comment?style=social)](https://github.com/belitre/argocd-diff-preview-pr-comment/stargazers)

A Go CLI application that processes ArgoCD application diffs from [argocd-diff-preview](https://github.com/dag-andersen/argocd-diff-preview), splits them by application, and posts organized comments on GitHub pull requests. This tool helps teams review ArgoCD changes more effectively by providing clear, application-specific diff summaries directly in PR comments.

## Features

- 🚀 Process ArgoCD application diffs
- 📦 Split diffs by application for better organization
- 💬 Post structured comments on GitHub PRs
- 🎯 Configurable log levels (debug, info, warn, error, fatal)
- 🔧 Built with Go for performance and reliability
- 📊 Comprehensive test coverage
- 🏗️ Cross-platform support (Linux, macOS, Windows)

## Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/belitre/argocd-diff-preview-pr-comment/releases).

Available architectures:
- Linux (amd64, arm64)
- macOS (arm64)
- Windows (amd64)

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

```bash
# Show version and help
argocd-diff-preview-pr-comment --help
argocd-diff-preview-pr-comment version

# Run with custom log level
argocd-diff-preview-pr-comment --log-level debug version

# Available log levels: debug, info, warn, error, fatal
```

## Configuration

The application requires the following environment variables:

- `GITHUB_TOKEN`: GitHub Personal Access Token with `repo` scope (for posting PR comments)

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

1. Make code changes following the [code style guidelines](CLAUDE.md#code-style)
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
├── Makefile                              # Build automation
├── go.mod                                # Go module definition
├── go.sum                                # Go module checksums
├── .releaserc.json                       # Semantic-release config
├── CHANGELOG.md                          # Auto-generated changelog
├── CLAUDE.md                             # Developer documentation
└── README.md                             # This file
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/amazing-feature`)
3. Make your changes following the code style guidelines
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