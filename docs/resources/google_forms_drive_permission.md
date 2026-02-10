---
page_title: "google_forms_drive_permission Resource"
description: |-
  Manages a Google Drive permission for a Drive-backed document (Forms/Sheets).
---

# google_forms_drive_permission

Manages a Google Drive permission for a Drive-backed document such as a Form or Spreadsheet.

## Example Usage

```hcl
resource "google_forms_spreadsheet" "example" {
  title = "Example Spreadsheet"
}

resource "google_forms_drive_permission" "share" {
  file_id = google_forms_spreadsheet.example.id

  type = "user"
  role = "writer"

  email_address = "user@example.com"
}
```

## Schema

### Required

- `file_id` (String) - Drive file ID.
- `type` (String) - Permission type: `user`, `group`, `domain`, or `anyone`.
- `role` (String) - Permission role: `reader`, `commenter`, or `writer`.

### Optional

- `email_address` (String) - Email for `user`/`group`.
- `domain` (String) - Domain for `domain`.
- `allow_file_discovery` (Boolean) - Discovery for `domain`/`anyone` (defaults to false).
- `send_notification_email` (Boolean) - Send notification email (defaults to false).
- `email_message` (String) - Optional email message.
- `supports_all_drives` (Boolean) - Shared drives support (defaults to false).

### Read-Only

- `id` (String) - Composite ID `fileID#permissionID`.
- `permission_id` (String) - Permission ID.
- `display_name` (String) - Display name (if available).

