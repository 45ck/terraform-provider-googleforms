terraform {
  required_providers {
    googleforms = {
      source  = "45ck/googleforms"
      version = "~> 0.1"
    }
  }
}

provider "googleforms" {}

variable "form_id" {
  type        = string
  description = "Existing Google Form ID to apply batchUpdate requests to."
}

# Escape hatch: apply raw Forms batchUpdate requests.
resource "googleforms_forms_batch_update" "update_info" {
  form_id = var.form_id

  # This JSON is an array of Forms Request objects.
  requests_json = jsonencode([
    {
      updateFormInfo = {
        info = {
          description = "Updated by googleforms_forms_batch_update"
        }
        updateMask = "description"
      }
    }
  ])

  include_form_in_response = false
  store_response_json      = false
}

