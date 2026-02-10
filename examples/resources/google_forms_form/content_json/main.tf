resource "googleforms_form" "advanced" {
  title               = "Advanced Survey (JSON Mode)"
  description         = "Uses content_json for unsupported question types."
  published           = true
  accepting_responses = true

  # Declarative JSON matching the Forms API items schema.
  # Use jsonencode() for plan-time validation and variable interpolation.
  content_json = jsonencode([
    {
      title = "Rate our service"
      questionItem = {
        question = {
          scaleQuestion = {
            low      = 1
            high     = 5
            lowLabel = "Poor"
            highLabel = "Excellent"
          }
        }
      }
    },
    {
      title = "Pick your favorites"
      questionItem = {
        question = {
          choiceQuestion = {
            type = "CHECKBOX"
            options = [
              { value = "Red" },
              { value = "Blue" },
              { value = "Green" }
            ]
          }
        }
      }
    }
  ])
}

output "advanced_form_url" {
  value = googleforms_form.advanced.responder_uri
}

