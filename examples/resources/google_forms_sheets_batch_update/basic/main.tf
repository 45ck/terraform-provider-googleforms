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

# Escape hatch: apply raw Sheets batchUpdate requests.
resource "googleforms_sheets_batch_update" "add_tab" {
  spreadsheet_id = googleforms_spreadsheet.example.id

  # This JSON is an array of Sheets Request objects.
  requests_json = jsonencode([
    {
      addSheet = {
        properties = {
          title = "RawTab"
        }
      }
    }
  ])

  include_spreadsheet_in_response = false
  store_response_json            = false
}


