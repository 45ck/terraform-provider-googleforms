terraform {
  required_providers {
    googleforms = {
      source  = "45ck/googleforms"
      version = "~> 0.1"
    }
  }
}

provider "googleforms" {}

resource "googleforms_spreadsheet" "example" {
  title = "Example Spreadsheet"
}

resource "googleforms_sheet" "config" {
  spreadsheet_id = googleforms_spreadsheet.example.id
  title          = "Config"
  row_count      = 50
  column_count   = 10
}


