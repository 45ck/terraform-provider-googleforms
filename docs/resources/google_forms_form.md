---
page_title: "google_forms_form Resource"
description: |-
  Manages a Google Form.
---

# google_forms_form

Manages a Google Form with full lifecycle support: create, read, update, delete, and import.

## Example Usage

### Basic Survey

```hcl
resource "google_forms_form" "survey" {
  title               = "Employee Survey"
  description         = "Annual employee satisfaction survey."
  published           = true
  accepting_responses = true

  item {
    item_key = "department"
    multiple_choice {
      question_text = "Which department are you in?"
      options       = ["Engineering", "Marketing", "Sales"]
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
```

### Quiz with Grading

```hcl
resource "google_forms_form" "quiz" {
  title               = "Geography Quiz"
  quiz                = true
  published           = true
  accepting_responses = true

  item {
    item_key = "q1"
    multiple_choice {
      question_text = "Capital of France?"
      options       = ["Paris", "London", "Rome"]
      required      = true
      grading {
        points         = 5
        correct_answer = "Paris"
      }
    }
  }
}
```

### JSON Escape Hatch

```hcl
resource "google_forms_form" "advanced" {
  title     = "Advanced Form"
  published = true

  content_json = jsonencode([
    {
      title = "Rate our service"
      questionItem = {
        question = {
          scaleQuestion = { low = 1, high = 5 }
        }
      }
    }
  ])
}
```

## Argument Reference

### Top-level

- `title` (String, Required) - The form title.
- `description` (String, Optional) - The form description.
- `published` (Boolean, Optional, Default: `false`) - Whether the form is published.
- `accepting_responses` (Boolean, Optional, Default: `false`) - Whether accepting responses. Requires `published = true`.
- `quiz` (Boolean, Optional, Default: `false`) - Enable quiz mode.
- `content_json` (String, Optional) - Declarative JSON array of form items. Mutually exclusive with `item` blocks. Use `jsonencode()`.

### `item` Block (Optional, Repeatable)

Mutually exclusive with `content_json`.

- `item_key` (String, Required) - Unique identifier for this item. Format: `[a-z][a-z0-9_]{0,63}`.

Exactly one of the following sub-blocks:

#### `multiple_choice` Sub-block

- `question_text` (String, Required) - The question text.
- `options` (List of String, Required) - Answer options (at least one).
- `required` (Boolean, Optional, Default: `false`)
- `grading` (Block, Optional) - Quiz grading. Requires `quiz = true`.

#### `short_answer` Sub-block

- `question_text` (String, Required)
- `required` (Boolean, Optional, Default: `false`)
- `grading` (Block, Optional)

#### `paragraph` Sub-block

- `question_text` (String, Required)
- `required` (Boolean, Optional, Default: `false`)
- `grading` (Block, Optional)

#### `grading` Sub-block

- `points` (Number, Required) - Point value.
- `correct_answer` (String, Optional) - Must match an option for multiple choice.
- `feedback_correct` (String, Optional)
- `feedback_incorrect` (String, Optional)

## Attribute Reference

- `id` - The Google Form ID.
- `responder_uri` - URL for respondents.
- `edit_uri` - URL to edit the form.
- `document_title` - The Google Drive document title.
- `google_item_id` (per item) - Google-assigned item ID.

## Import

```bash
terraform import google_forms_form.example FORM_ID
```

Items receive auto-generated `item_key` values (`item_0`, `item_1`, ...).
Review and rename these in your configuration after import.

## Notes

- Forms created via API are unpublished by default.
- `published` must be `true` before `accepting_responses` can be `true`.
- `content_json` and `item` blocks are mutually exclusive.
- `content_json` mode has no external drift detection (hash-based only).
