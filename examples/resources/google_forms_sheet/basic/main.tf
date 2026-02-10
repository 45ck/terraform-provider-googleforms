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

resource "google_forms_sheet" "config" {
  spreadsheet_id = google_forms_spreadsheet.example.id
  title          = "Config"
  row_count      = 50
  column_count   = 10
}

