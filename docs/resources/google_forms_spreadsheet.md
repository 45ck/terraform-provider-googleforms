---
page_title: "google_forms_spreadsheet Resource"
description: |-
  Manages a Google Sheets spreadsheet.
---

# google_forms_spreadsheet

Manages a Google Sheets spreadsheet (Drive-backed document).

## Example Usage

```hcl
resource "google_forms_spreadsheet" "example" {
  title = "Example Spreadsheet"
}
```

## Schema

### Required

- `title` (String) - Spreadsheet title.

### Optional

- `locale` (String) - Spreadsheet locale.
- `time_zone` (String) - Spreadsheet time zone.

### Read-Only

- `id` (String) - Spreadsheet ID.
- `url` (String) - Spreadsheet URL.

