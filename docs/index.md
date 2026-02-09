---
page_title: "Google Forms Provider"
description: |-
  Manage Google Forms as infrastructure using Terraform.
---

# Google Forms Provider

The Google Forms provider allows you to create, update, and delete Google Forms
using Terraform. It supports multiple question types, quiz mode with grading,
publish state control, and a JSON escape hatch for advanced use cases.

## Authentication

The provider supports three authentication methods (in priority order):

1. **`credentials` attribute** in the provider block
2. **`GOOGLE_CREDENTIALS` environment variable**
3. **Application Default Credentials (ADC)** - recommended

Required OAuth scopes: `forms.body`, `drive.file`

### Using Application Default Credentials (Recommended)

```bash
gcloud auth application-default login --scopes=\
https://www.googleapis.com/auth/forms.body,\
https://www.googleapis.com/auth/drive.file
```

### Using a Service Account

```bash
export GOOGLE_CREDENTIALS=/path/to/service-account.json
```

## Example Usage

```hcl
provider "googleforms" {}

resource "google_forms_form" "example" {
  title               = "My Form"
  published           = true
  accepting_responses = true

  item {
    item_key = "q1"
    short_answer {
      question_text = "What is your name?"
      required      = true
    }
  }
}
```

## Schema

### Optional

- `credentials` (String, Sensitive) - Service account JSON key content or file path.
- `impersonate_user` (String) - Email of user to impersonate via domain-wide delegation.
