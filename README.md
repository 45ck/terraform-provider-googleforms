# Terraform Provider for Google Forms

[![Tests](https://github.com/45ck/terraform-provider-googleforms/actions/workflows/test.yml/badge.svg)](https://github.com/45ck/terraform-provider-googleforms/actions/workflows/test.yml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A Terraform provider for managing Google Forms as infrastructure, with optional Google Drive and Google Sheets helpers for common workflows (folder placement, permissions, response spreadsheets, etc.).

## What You Can Manage

This provider is intentionally "typed-first, escape-hatch always":

- Use strongly-typed HCL for common Forms/Sheets/Drive operations.
- When Google adds API features faster than provider schema can evolve, use `*_batch_update` resources for full-power raw JSON requests.

### Resources

| Area | Resource | Purpose |
|------|----------|---------|
| Forms | `googleforms_form` | Typed Form + items (questions), quiz, publish/accept responses, Drive folder placement |
| Forms | `googleforms_forms_batch_update` | Escape hatch for Forms `forms.batchUpdate` |
| Forms | `googleforms_response_sheet` | Track/validate Form <-> Spreadsheet association |
| Sheets | `googleforms_spreadsheet` | Spreadsheet + Drive folder placement |
| Sheets | `googleforms_sheet` | Sheet/tab within a spreadsheet |
| Sheets | `googleforms_sheet_values` | Bounded A1 range writes, optional read-back drift detection |
| Sheets | `googleforms_sheets_batch_update` | Escape hatch for Sheets `spreadsheets.batchUpdate` |
| Sheets | `googleforms_sheets_named_range` | Named ranges via batchUpdate |
| Sheets | `googleforms_sheets_protected_range` | Protected ranges via batchUpdate |
| Sheets | `googleforms_sheets_developer_metadata` | Developer metadata via batchUpdate |
| Sheets | `googleforms_sheets_data_validation` | Data validation rules (JSON) via batchUpdate |
| Sheets | `googleforms_sheets_conditional_format_rule` | Conditional format rule (JSON) addressed by index |
| Drive | `googleforms_drive_folder` | Create/manage Drive folders |
| Drive | `googleforms_drive_file` | Adopt existing file; rename/move; optional delete-on-destroy |
| Drive | `googleforms_drive_permission` | Share Drive-backed documents |

### Data Sources

| Area | Data source | Purpose |
|------|-------------|---------|
| Forms | `data.googleforms_form` | Read a Form by ID |
| Drive | `data.googleforms_drive_file` | Read a Drive file by ID |
| Sheets | `data.googleforms_spreadsheet` | Read a spreadsheet by ID |
| Sheets | `data.googleforms_sheet_values` | Read sheet values for an A1 range |

## Documentation

- Provider docs: `docs/`
- Resources: `docs/resources/`
- Data sources: `docs/data-sources/`
- Guide: `docs/guides/import-existing-form.md`

## Quick Start

### Prerequisites

1. A Google Cloud project with the **Forms API** enabled
2. A Google Cloud project with the **Drive API** enabled (required for folder placement, permissions, and some document operations)
3. A Google Cloud project with the **Sheets API** enabled (required for spreadsheet resources)
4. A service account with appropriate permissions
5. Terraform >= 1.0

### Provider Configuration

```hcl
terraform {
  required_providers {
    googleforms = {
      source  = "45ck/googleforms"
      version = "~> 0.1"
    }
  }
}

# Uses Application Default Credentials by default.
# Set GOOGLE_CREDENTIALS env var for service account JSON.
provider "googleforms" {}
```

### Example: Simple Survey

```hcl
resource "googleforms_form" "survey" {
  title               = "Employee Satisfaction Survey"
  description         = "Managed by Terraform."
  published           = true
  accepting_responses = true
  update_strategy     = "targeted" # avoids replace-all when only editing existing items

  item {
    item_key = "department"
    multiple_choice {
      question_text = "Which department are you in?"
      options       = ["Engineering", "Marketing", "Sales", "HR"]
      required      = true
    }
  }

  item {
    item_key = "name"
    short_answer {
      question_text = "What is your name?"
      required      = true
    }
  }

  item {
    item_key = "feedback"
    paragraph {
      question_text = "Any additional feedback?"
    }
  }
}

output "form_url" {
  value = googleforms_form.survey.responder_uri
}
```

### Example: Quiz with Grading

```hcl
resource "googleforms_form" "quiz" {
  title               = "Geography Quiz"
  quiz                = true
  published           = true
  accepting_responses = true

  item {
    item_key = "q1_capital"
    multiple_choice {
      question_text = "What is the capital of France?"
      options       = ["Paris", "London", "Rome", "Berlin"]
      required      = true
      grading {
        points         = 5
        correct_answer = "Paris"
      }
    }
  }
}
```

### Example: JSON Escape Hatch

For question types or settings not yet natively supported, use `content_json`:

```hcl
resource "googleforms_form" "advanced" {
  title     = "Advanced Form"
  published = true

  content_json = jsonencode([
    {
      title = "Rate our service"
      questionItem = {
        question = {
          scaleQuestion = {
            low      = 1
            high     = 5
            lowLabel = "Poor"
            highLabel = "Excellent"
          }
        }
      }
    }
  ])
}
```

## Form Management Model (Important)

`googleforms_form` supports two complementary item modes:

- Typed items via `item { ... }` blocks for supported question/item types.
- Raw JSON items via `content_json` (mutually exclusive with `item` blocks).

When using typed `item` blocks:

- `manage_mode = "all"`: treat the configured item list as authoritative for the whole form.
- `manage_mode = "partial"`: only manage the configured items (by `item_key`) and leave other items untouched.

When changing items, choose an update strategy:

- `update_strategy = "targeted"`: uses Forms `batchUpdate` to update/move/create/delete items correlated by `item_key` + stored `google_item_id`. Safer for preserving response mappings and external integrations. Refuses question type changes.
- `update_strategy = "replace_all"`: deletes/recreates items when changes occur. This can break response mappings/integrations; gated by `dangerously_replace_all_items`.

Branching forms (go-to-section behavior) are supported for choice options via `option { ... }`. When you reference sections by `go_to_section_key`, the provider resolves keys to IDs after item IDs exist.

## Authentication

The provider supports three authentication methods (in priority order):

1. **`credentials` attribute** — Path to or content of a service account JSON key
2. **`GOOGLE_CREDENTIALS` environment variable** — Path to service account JSON
3. **Application Default Credentials** — Automatically detected from environment

Optional: `impersonate_user` for Google Workspace domain-wide delegation.

Required OAuth scopes: `forms.body`, `drive.file`, `spreadsheets`.

## Limitations / Gotchas

These are the top "surprises" users hit when automating Forms and Drive-backed docs:

- File upload questions: the Forms API does not support creating them via the same typed item workflows. The provider supports `file_upload` primarily for imported/existing items and state.
- Response destination linking: the Forms REST API does not support programmatically linking a Form to a response Spreadsheet. `googleforms_response_sheet` tracks and can validate the association, but cannot create it.
- `revision_id` write control: when using `conflict_policy = "fail"`, the `revision_id` is only valid for a limited time (Google currently documents ~24 hours). Plan/apply long after the last read may require a refresh.
- `googleforms_sheets_conditional_format_rule` uses an index into `conditionalFormats`. Out-of-band edits that insert/remove rules can shift indexes and cause unexpected diffs.
- `googleforms_sheet_values` is intentionally range-scoped to prevent state explosion. Manage large sheets as many small ranges (or use `googleforms_sheets_batch_update`).

## Importing Existing Forms

```bash
terraform import googleforms_form.existing FORM_ID
```

See the [import guide](docs/guides/import-existing-form.md) for details.

## Examples

See `examples/` for complete configurations:

- Forms: `examples/resources/google_forms_form/`
- Sheets: `examples/resources/google_forms_spreadsheet/`, `examples/resources/google_forms_sheet/`, `examples/resources/google_forms_sheet_values/`
- Escape hatches: `examples/resources/google_forms_forms_batch_update/`, `examples/resources/google_forms_sheets_batch_update/`
- Drive: `examples/resources/google_forms_drive_folder/`, `examples/resources/google_forms_drive_permission/`

## Development

```bash
# Install dependencies
go mod download

# Install pre-commit hooks
pre-commit install

# Build
make build

# Run all quality checks
make ci

# Run unit tests
make test

# Run acceptance tests (requires GOOGLE_CREDENTIALS)
make test-acc
```

Windows note: if `go test` fails with `Access is denied` due to temp exe execution restrictions, use `make test-docker` and `make docs-docker`.

See [CONTRIBUTING.md](CONTRIBUTING.md) for full development guide.

## Scope / Non-Goals (For Now)

- Form responses export/management is not implemented yet (responses are usually treated as data, not infrastructure).
- Apps Script automation is not included in this provider core (service-account-only CI constraints and operational complexity).

## License

Apache License 2.0. See [LICENSE](LICENSE).

