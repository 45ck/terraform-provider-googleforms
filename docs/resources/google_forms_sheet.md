---
page_title: "google_forms_sheet Resource"
description: |-
  Manages a sheet (tab) within a Google Sheets spreadsheet.
---

# google_forms_sheet

Manages an individual sheet (tab) within a Google Spreadsheet.

## Example Usage

```hcl
resource "google_forms_spreadsheet" "example" {
  title = "Example Spreadsheet"
}

resource "google_forms_sheet" "config" {
  spreadsheet_id = google_forms_spreadsheet.example.id
  title          = "Config"
  row_count      = 50
  column_count   = 10
}
```

## Schema

### Required

- `spreadsheet_id` (String) - Spreadsheet ID.
- `title` (String) - Sheet title.

### Optional

- `row_count` (Number) - Row count (defaults to 1000).
- `column_count` (Number) - Column count (defaults to 26).

### Read-Only

- `id` (String) - Composite ID `spreadsheetID#sheetID`.
- `sheet_id` (Number) - Numeric sheet ID.
- `index` (Number) - Zero-based position in the spreadsheet.

