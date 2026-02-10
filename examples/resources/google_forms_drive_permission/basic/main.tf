terraform {
  required_providers {
    googleforms = {
      source  = "45ck/googleforms"
      version = "~> 0.1"
    }
  }
}

provider "googleforms" {}

resource "google_forms_spreadsheet" "example" {
  title = "Example Spreadsheet"
}

# Share the spreadsheet with a user.
resource "google_forms_drive_permission" "share" {
  file_id = google_forms_spreadsheet.example.id

  type = "user"
  role = "writer"

  email_address = "user@example.com"

  send_notification_email = false
  supports_all_drives     = false
}

