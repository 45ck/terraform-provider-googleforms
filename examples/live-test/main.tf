terraform {
  required_providers {
    googleforms = { source = "45ck/googleforms" }
  }
}

provider "googleforms" {}

resource "google_forms_form" "test" {
  title       = "Terraform Test Form"
  description = "Created by Terraform provider live test."

  item {
    item_key = "q1"
    short_answer {
      question_text = "What is your name?"
      required      = true
    }
  }
}

output "form_url" { value = google_forms_form.test.responder_uri }
output "edit_url" { value = google_forms_form.test.edit_uri }
