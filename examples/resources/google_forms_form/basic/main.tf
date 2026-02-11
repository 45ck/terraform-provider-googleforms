resource "googleforms_form" "survey" {
  title               = "Employee Satisfaction Survey"
  description         = "Managed by Terraform."
  published           = true
  accepting_responses = true
  email_collection_type = "VERIFIED"

  item {
    item_key = "department"
    multiple_choice {
      question_text = "Which department are you in?"
      options       = ["Engineering", "Marketing", "Sales", "HR"]
      required      = true
    }
  }

  item {
    item_key = "skills"
    multiple_choice_grid {
      question_text = "How do you feel about the following?"
      rows          = ["Work-life balance", "Role clarity", "Manager support"]
      columns       = ["Needs improvement", "OK", "Great"]
      required      = true
    }
  }

  item {
    item_key = "name"
    short_answer {
      question_text = "What is your name?"
      required      = true
    }
  }

  item {
    item_key = "feedback"
    paragraph {
      question_text = "Any additional feedback?"
    }
  }
}

output "form_url" {
  value = googleforms_form.survey.responder_uri
}

output "form_id" {
  value = googleforms_form.survey.id
}

