# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Expanded Forms resource coverage:
  - New item types: dropdown, checkbox, date, date_time, scale, time, rating, section_header, text_item, image, video
  - Grid questions: multiple_choice_grid, checkbox_grid
  - Email collection settings (DO_NOT_COLLECT, VERIFIED, RESPONDER_INPUT)
  - Branching/navigation via choice option `go_to_*` settings, plus `shuffle` and `has_other` support
- Forms management capabilities:
  - Partial item management mode and new-item placement policy
  - Conflict detection using form revision_id (optional fail-on-drift write control)
  - Targeted item updates strategy (batchUpdate-based) and structural move/insert handling
- Escape hatches:
  - `googleforms_forms_batch_update` for raw Forms `forms.batchUpdate` requests
  - `googleforms_sheets_batch_update` for raw Sheets `spreadsheets.batchUpdate` requests
- Sheets resources and helpers:
  - `googleforms_spreadsheet`, `googleforms_sheet`, `googleforms_sheet_values`
  - Named ranges, protected ranges, developer metadata
  - Data validation and conditional format rule helpers (JSON-driven)
- Drive resources for Drive-backed documents:
  - `googleforms_drive_folder`, `googleforms_drive_file`, `googleforms_drive_permission`
- New data sources:
  - `data.googleforms_form`, `data.googleforms_drive_file`, `data.googleforms_spreadsheet`, `data.googleforms_sheet_values`
- Documentation improvements:
  - Clear “What you can manage” tables in README and better docs entrypoints
  - Import guide expanded for additional resources

### Changed
- Improved docs generation enforcement and cross-platform contributor experience:
  - Added Docker-based `make test-docker` and `make docs-docker` targets
  - `make ci` now runs the full quality suite consistently
  - Added `.gitattributes` to normalize line endings

### Fixed
- Removed local path references from review documentation.

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
