# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-02-09

### Added
- `googleforms_form` resource with full CRUD lifecycle
- HCL item blocks for multiple_choice, short_answer, and paragraph question types
- `content_json` attribute for declarative JSON-based form item definition
- Quiz mode with grading support (points, correct answers, feedback)
- Publish settings management (published, accepting_responses)
- Plan modifier for semantic JSON diff suppression on content_json
- Seven config validators: mutual exclusivity, publish guards, unique item keys,
  exactly-one sub-block, options required for choice, correct answer in options,
  grading requires quiz
- Google Forms API client with Create, Get, BatchUpdate, Delete, and SetPublishSettings
- Google Drive API client for form deletion (trash to Drive)
- Comprehensive unit tests for all CRUD operations and validators
- Mock-based test infrastructure with testutil helpers
- Initial repository structure with quality infrastructure
- golangci-lint v2 strict configuration
- Pre-commit hooks (gofumpt, goimports, commitlint)
- CI workflows: test (9 gates), acceptance, release
- GNUmakefile with all quality targets
- Quality tools: file limits, coverage, coupling, mutation testing
- Apache 2.0 license
- Contributing guide, Code of Conduct, Security policy
