terraform {
  required_providers {
    googleforms = {
      source  = "45ck/googleforms"
      version = "~> 0.1"
    }
  }
}

provider "googleforms" {}

data "google_forms_spreadsheet" "example" {
  id = "spreadsheet-id"
}

