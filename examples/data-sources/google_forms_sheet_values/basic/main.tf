terraform {
  required_providers {
    googleforms = {
      source  = "45ck/googleforms"
      version = "~> 0.1"
    }
  }
}

provider "googleforms" {}

data "googleforms_sheet_values" "example" {
  spreadsheet_id = "spreadsheet-id"
  range          = "Config!A1:D20"
}


