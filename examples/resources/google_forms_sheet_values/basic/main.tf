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
}

resource "googleforms_sheet_values" "config_values" {
  spreadsheet_id = googleforms_spreadsheet.example.id
  range          = "Config!A1:B2"

  value_input_option = "RAW"
  read_back          = true

  rows = [
    {
      cells = ["key", "value"]
    },
    {
      cells = ["foo", "bar"]
    },
  ]
}


