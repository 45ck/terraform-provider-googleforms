default: build

BINARY=terraform-provider-googleforms
GOFLAGS=-trimpath
LDFLAGS=-s -w
COVERAGE_THRESHOLD=85
PACKAGE_COVERAGE_THRESHOLD=75
MUTATION_THRESHOLD=60

.PHONY: build install clean fmt lint test test-race test-acc coverage \
        mutation coupling security docs ci help

build: ## Build the provider binary
	go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BINARY)

install: build ## Install the provider locally
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/hashicorp/googleforms/0.1.0/$$(go env GOOS)_$$(go env GOARCH)
	cp $(BINARY) ~/.terraform.d/plugins/registry.terraform.io/hashicorp/googleforms/0.1.0/$$(go env GOOS)_$$(go env GOARCH)/

clean: ## Remove build artifacts
	rm -f $(BINARY)
	rm -f coverprofile.txt coverage.html

fmt: ## Format code with gofumpt and goimports
	gofumpt -w .
	goimports -w .

lint: ## Run golangci-lint
	golangci-lint run

test: ## Run unit tests
	go test ./... -short -count=1

test-race: ## Run unit tests with race detector
	go test -race ./... -short -count=1

test-acc: ## Run acceptance tests (requires GOOGLE_CREDENTIALS)
	TF_ACC=1 go test ./internal/... -v -timeout 30m -count=1

coverage: ## Run tests with coverage and enforce thresholds
	go test ./... -short -coverprofile=coverprofile.txt -covermode=atomic
	go tool cover -html=coverprofile.txt -o coverage.html
	@./tools/quality/check_coverage.sh

mutation: ## Run mutation testing on changed packages
	@./tools/quality/mutation.sh

coupling: ## Check package coupling thresholds
	@go run ./tools/quality/check_coupling.go

security: ## Run security scanners
	govulncheck ./...

docs: ## Generate and validate provider documentation
	@echo "Generating docs with tfplugindocs..."
	@which tfplugindocs > /dev/null 2>&1 && tfplugindocs generate --provider-name googleforms || echo "tfplugindocs not installed; skipping"

file-limits: ## Check file size and shape limits
	@go run ./tools/quality/check_file_limits.go

sweep: ## Run test sweeper to clean up orphaned resources
	@echo "WARNING: This will delete tf-test-* forms from the test account"
	go test ./internal/testutil/... -v -sweep=all -timeout 10m

ci: fmt lint test test-race coverage security file-limits ## Run all CI checks locally
	@echo "All CI checks passed."

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
