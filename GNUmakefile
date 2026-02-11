default: build

BINARY=terraform-provider-googleforms
GOFLAGS=-trimpath
LDFLAGS=-s -w
COVERAGE_THRESHOLD=85
PACKAGE_COVERAGE_THRESHOLD=75
MUTATION_THRESHOLD=60
DOCKER_GO_IMAGE=golang:1.24

.PHONY: build install clean fmt lint test test-race test-acc coverage \
        mutation coupling security docs docs-docker test-docker ci ci-fast help

build: ## Build the provider binary
	go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BINARY)

install: build ## Install the provider locally
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/45ck/googleforms/0.1.0/$$(go env GOOS)_$$(go env GOARCH)
	cp $(BINARY) ~/.terraform.d/plugins/registry.terraform.io/45ck/googleforms/0.1.0/$$(go env GOOS)_$$(go env GOARCH)/

clean: ## Remove build artifacts
	rm -f $(BINARY)
	rm -f coverprofile.txt coverage.html

fmt: ## Format code with gofumpt and goimports
	# Only format module packages (avoid touching local caches like .cache/).
	gofumpt -w $$(go list -f '{{.Dir}}' ./...)
	goimports -w $$(go list -f '{{.Dir}}' ./...)

lint: ## Run golangci-lint
	golangci-lint run

test: ## Run unit tests
	go test ./... -short -count=1

test-race: ## Run unit tests with race detector
	go test -race ./... -short -count=1

test-docker: ## Run unit tests in Docker (useful on Windows hosts with temp exe restrictions)
	docker run --rm -v "$$PWD":/src -w /src $(DOCKER_GO_IMAGE) bash -lc \
		'export PATH=/usr/local/go/bin:$$PATH; go test ./... -short -count=1'

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

docs: ## Generate and validate provider documentation (fails if docs are out of date)
	@command -v tfplugindocs > /dev/null 2>&1 || ( \
		echo "tfplugindocs not installed."; \
		echo "Install: go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest"; \
		echo "Or run: make docs-docker"; \
		exit 1 \
	)
	tfplugindocs generate --provider-name googleforms
	git diff --exit-code
	tfplugindocs validate --provider-name googleforms

docs-docker: ## Generate/validate docs in Docker
	docker run --rm -v "$$PWD":/src -w /src $(DOCKER_GO_IMAGE) bash -lc \
		'export PATH=/usr/local/go/bin:/go/bin:$$PATH; \
		go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest; \
		tfplugindocs generate --provider-name googleforms; \
		git diff --exit-code; \
		tfplugindocs validate --provider-name googleforms'

file-limits: ## Check file size and shape limits
	@go run ./tools/quality/check_file_limits.go

sweep: ## Run test sweeper to clean up orphaned resources
	@echo "WARNING: This will delete tf-test-* forms from the test account"
	go test ./internal/testutil/... -v -sweep=all -timeout 10m

ci: fmt lint test test-race coverage mutation coupling security file-limits docs ## Run all CI checks locally
	@echo "All CI checks passed."

ci-fast: fmt lint test coverage security file-limits ## Run a faster local check suite (skips mutation/coupling/docs)
	@echo "Fast CI checks passed."

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
