.PHONY: build build-cross test fmt lint coverage clean help tidy all install-semantic-release package dist release release-dry-run

# Go parameters
GOCMD=/usr/local/go/bin/go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
GOMOD=$(GOCMD) mod
BINARY_NAME=argocd-diff-preview-pr-comment
CMD_PATH=./cmd/argocd-diff-preview-pr-comment
BUILD_DIR=./build
VERSION?=canary
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X github.com/belitre/argocd-diff-preview-pr-comment/pkg/version.Version=$(VERSION) \
	-X github.com/belitre/argocd-diff-preview-pr-comment/pkg/version.Commit=$(COMMIT)"

# Build targets
PLATFORMS=linux/amd64 linux/arm64 darwin/arm64 windows/amd64

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build application for current platform only
	@echo "Building for local platform..."
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete! Binary: $(BUILD_DIR)/$(BINARY_NAME)"

build-cross: ## Build application for multiple architectures (Linux-amd64, Linux-arm64, Darwin-arm64, Windows-amd64)
	@echo "Building for multiple platforms..."
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		platform_dir=$(BUILD_DIR)/$$GOOS-$$GOARCH; \
		mkdir -p $$platform_dir; \
		output_name=$$platform_dir/$(BINARY_NAME); \
		if [ $$GOOS = "windows" ]; then \
			output_name=$$output_name.exe; \
		fi; \
		echo "Building $$output_name..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH $(GOBUILD) $(LDFLAGS) -o $$output_name $(CMD_PATH); \
		if [ $$? -ne 0 ]; then \
			echo "Failed to build for $$platform"; \
			exit 1; \
		fi; \
	done
	@echo "Build complete! Binaries are in $(BUILD_DIR)/*/"

test: ## Run Go tests with verbose output
	@echo "Running tests..."
	$(GOTEST) -v ./...

fmt: ## Format code using go fmt
	@echo "Formatting code..."
	$(GOFMT) ./...

lint: ## Run Go linter (go vet)
	@echo "Running linter..."
	$(GOVET) ./...
	@echo "Lint check passed!"

coverage: ## Generate and display test coverage report
	@echo "Generating coverage report..."
	@mkdir -p $(BUILD_DIR)
	$(GOTEST) -v -coverprofile=$(BUILD_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report generated: $(BUILD_DIR)/coverage.html"
	$(GOCMD) tool cover -func=$(BUILD_DIR)/coverage.out

clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete!"

tidy: ## Run go mod tidy
	@echo "Running go mod tidy..."
	$(GOMOD) tidy

all: fmt lint test build-cross ## Run fmt, lint, test, and build-cross

install-semantic-release: ## Install semantic-release dependencies
	@echo "Installing semantic-release and plugins..."
	npm install -g \
		semantic-release@latest \
		@semantic-release/git@latest \
		@semantic-release/changelog@latest \
		@semantic-release/exec@latest \
		conventional-changelog-conventionalcommits@latest
	@echo "Semantic-release installed successfully!"

package: ## Package binaries as tar.gz and zip archives
	@echo "Packaging binaries..."
	@echo "Version: $(VERSION)"
	@cd $(BUILD_DIR) && \
	for dir in */; do \
		if [ -d "$$dir" ]; then \
			arch_name=$${dir%/}; \
			archive_name=$(BINARY_NAME)-$(VERSION)-$$arch_name; \
			echo "Packaging $$arch_name as $$archive_name..."; \
			tar -czf "$$archive_name.tar.gz" "$$arch_name"; \
			zip -qr "$$archive_name.zip" "$$arch_name"; \
		fi; \
	done
	@echo "Packaging complete! Archives are in $(BUILD_DIR)/"
	@ls -lh $(BUILD_DIR)/*.tar.gz $(BUILD_DIR)/*.zip 2>/dev/null || true

dist: build-cross package ## Build for all platforms and create distribution archives
	@echo "Distribution build complete!"

release: ## Run semantic-release to create a new release
	@echo "Running semantic-release..."
	npx semantic-release

release-dry-run: ## Run semantic-release in dry-run mode (no actual release)
	@echo "Running semantic-release in dry-run mode..."
	npx semantic-release --dry-run
