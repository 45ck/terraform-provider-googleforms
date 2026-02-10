---
page_title: "google_forms_sheet_values Resource"
description: |-
  Manages a bounded A1 range of values in a Google Sheets spreadsheet.
---

# google_forms_sheet_values

Manages a bounded range of values in a spreadsheet using A1 notation.

## Example Usage

```hcl
resource "google_forms_spreadsheet" "example" {
  title = "Example Spreadsheet"
}

resource "google_forms_sheet" "config" {
  spreadsheet_id = google_forms_spreadsheet.example.id
  title          = "Config"
}

resource "google_forms_sheet_values" "config_values" {
  spreadsheet_id = google_forms_spreadsheet.example.id
  range          = "Config!A1:B2"

  value_input_option = "RAW"
  read_back          = true

  rows = [
    { cells = ["key", "value"] },
    { cells = ["foo", "bar"] },
  ]
}
```

## Schema

### Required

- `spreadsheet_id` (String) - Spreadsheet ID.
- `range` (String) - A1 range to write.
- `rows` (List of Object) - Rows to write (`cells` is a list of strings).

### Optional

- `value_input_option` (String) - `RAW` or `USER_ENTERED` (defaults to `RAW`).
- `read_back` (Boolean) - If true, reads values back during Read to detect drift (defaults to true).

### Read-Only

- `id` (String) - Composite ID `spreadsheetID#A1_RANGE`.
- `updated_range` (String) - API-reported updated range.

