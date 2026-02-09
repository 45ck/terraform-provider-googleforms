# Terraform Provider for Google Forms

[![Tests](https://github.com/45ck/terraform-provider-googleforms/actions/workflows/test.yml/badge.svg)](https://github.com/45ck/terraform-provider-googleforms/actions/workflows/test.yml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A Terraform provider for managing Google Forms as infrastructure. Create, update, and delete Google Forms with full lifecycle management using Terraform.

## Features

- Create and manage Google Forms with HCL configuration
- Support for multiple choice, short answer, and paragraph questions
- Quiz mode with grading (correct answers, point values, feedback)
- Control publish state and response acceptance
- Raw JSON escape hatch for unsupported question types
- Import existing Google Forms into Terraform
- Service account and Application Default Credentials authentication

## Quick Start

### Prerequisites

1. A Google Cloud project with the **Forms API** and **Drive API** enabled
2. A service account with appropriate permissions
3. Terraform >= 1.0

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
resource "google_forms_form" "survey" {
  title               = "Employee Satisfaction Survey"
  description         = "Managed by Terraform."
  published           = true
  accepting_responses = true

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
  value = google_forms_form.survey.responder_uri
}
```

### Example: Quiz with Grading

```hcl
resource "google_forms_form" "quiz" {
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

For question types not yet natively supported:

```hcl
resource "google_forms_form" "advanced" {
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

## Authentication

The provider supports three authentication methods (in priority order):

1. **`credentials` attribute** — Path to or content of a service account JSON key
2. **`GOOGLE_CREDENTIALS` environment variable** — Path to service account JSON
3. **Application Default Credentials** — Automatically detected from environment

Required OAuth scopes: `forms.body`, `drive.file`

## Importing Existing Forms

```bash
terraform import google_forms_form.existing FORM_ID
```

See the [import guide](docs/guides/import-existing-form.md) for details.

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

See [CONTRIBUTING.md](CONTRIBUTING.md) for full development guide.

## Roadmap

| Version | Status | Features |
|---------|--------|----------|
| v0.1.0 | In Progress | MVP: 3 question types, quiz, publish state, JSON mode, import |
| v0.2.0 | Planned | All question types, Drive folder placement |
| v0.3.0 | Planned | Permissions management, targeted updates |
| v1.0.0 | Planned | Stable release, full Phase 2 features |

## License

Apache License 2.0. See [LICENSE](LICENSE).
