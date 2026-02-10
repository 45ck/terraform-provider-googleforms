terraform {
  required_providers {
    googleforms = {
      source  = "45ck/googleforms"
      version = "~> 0.1"
    }
  }
}

provider "googleforms" {}

resource "googleforms_drive_folder" "example" {
  name = "tf-example-googleforms-folder"
}

output "folder_id" {
  value = googleforms_drive_folder.example.id
}

