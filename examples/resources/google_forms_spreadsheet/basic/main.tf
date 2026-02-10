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


