terraform {
  required_providers {
    googleforms = {
      source  = "45ck/googleforms"
      version = "~> 0.1"
    }
  }
}

provider "googleforms" {}

variable "file_id" {
  type        = string
  description = "Existing Drive file ID to manage."
}

resource "googleforms_drive_file" "example" {
  file_id = var.file_id

  # Optional: rename/move the file.
  # name      = "New Name"
  # folder_id = "FOLDER_ID"

  delete_on_destroy = false
}

