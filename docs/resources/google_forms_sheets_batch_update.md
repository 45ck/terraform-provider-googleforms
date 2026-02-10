---
page_title: "google_forms_sheets_batch_update Resource"
description: |-
  Escape hatch for applying raw Sheets batchUpdate requests.
---

# google_forms_sheets_batch_update

Escape hatch resource for applying raw `spreadsheets.batchUpdate` requests.

This resource is imperative: it re-applies requests on Create/Update and does not attempt to reconstruct full desired state from the API.

## Example Usage

```hcl
resource "google_forms_spreadsheet" "example" {
  title = "Example Spreadsheet"
}

resource "google_forms_sheets_batch_update" "add_tab" {
  spreadsheet_id = google_forms_spreadsheet.example.id

  requests_json = jsonencode([
    {
      addSheet = {
        properties = {
          title = "RawTab"
        }
      }
    }
  ])
}
```

## Schema

### Required

- `spreadsheet_id` (String) - Spreadsheet ID.
- `requests_json` (String) - JSON array of Sheets `Request` objects, or a JSON object containing `requests`.

### Optional

- `include_spreadsheet_in_response` (Boolean) - Include spreadsheet in response (defaults to false).
- `store_response_json` (Boolean) - Store response JSON in state (defaults to false).

### Read-Only

- `id` (String) - Deterministic ID derived from inputs.
- `response_json` (String) - Response JSON when `store_response_json` is true.

