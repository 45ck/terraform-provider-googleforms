# Contributing to terraform-provider-googleforms

Thank you for your interest in contributing! This document provides guidelines
and instructions for contributing to this project.

## Development Setup

### Prerequisites

- [Go](https://golang.org/dl/) >= 1.22
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [golangci-lint](https://golangci-lint.run/usage/install/) v2
- [gofumpt](https://github.com/mvdan/gofumpt)
- [pre-commit](https://pre-commit.com/#install) (recommended)

### Getting Started

```bash
git clone https://github.com/45ck/terraform-provider-googleforms.git
cd terraform-provider-googleforms
go mod download
pre-commit install
make build
```

### Running Tests

```bash
# Unit tests (no credentials required)
make test

# Unit tests with race detector
make test-race

# Acceptance tests (requires GOOGLE_CREDENTIALS env var)
export GOOGLE_CREDENTIALS=/path/to/service-account.json
make test-acc
```

### Code Quality

This project enforces strict quality gates. All checks must pass before merging:

| Gate | Command | Description |
|------|---------|-------------|
| Format | `make fmt` | gofumpt + goimports |
| Lint | `make lint` | golangci-lint v2 (strict config) |
| Unit Tests | `make test` | All unit tests |
| Race Detection | `make test-race` | Tests with Go race detector |
| Coverage | `make coverage` | >= 85% total, >= 75% per package |
| Mutation Testing | `make mutation` | >= 60% on changed packages |
| Coupling | `make coupling` | Package fan-in/fan-out thresholds |
| Security | `make security` | govulncheck |
| Docs | `make docs` | tfplugindocs validation |

Run all checks locally with: `make ci`

### Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add checkbox question type support
fix: handle 404 on form deletion gracefully
docs: add import workflow guide
test: add acceptance tests for quiz grading
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`

### File Limits

- Max **400 lines** per `.go` file (excluding generated code)
- Max **10 exported functions/types** per file
- Max cyclomatic complexity: **10**
- Max cognitive complexity: **15**
- Max function length: **60 lines / 40 statements**
- Max line length: **120 characters**

### Pull Request Process

1. Fork the repository and create a feature branch
2. Write tests for your changes
3. Ensure all quality gates pass (`make ci`)
4. Submit a pull request with a clear description
5. Wait for review from CODEOWNERS (2 approvals required)

### Branch Protection

The `main` branch has strict protection:

- Pull request required (no direct pushes)
- 2 approvals required + CODEOWNERS review
- All 9 CI status checks must pass
- Conversations must be resolved

## Acceptance Tests

Acceptance tests create real Google Forms and require credentials:

1. Create a GCP project with Forms API and Drive API enabled
2. Create a service account with domain-wide delegation (if Workspace)
3. Set `GOOGLE_CREDENTIALS` to the service account JSON path
4. Run `make test-acc`

Test forms are prefixed with `tf-test-` for cleanup identification.

## License

By contributing, you agree that your contributions will be licensed under
the Apache License 2.0.
