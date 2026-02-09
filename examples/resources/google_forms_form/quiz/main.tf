resource "google_forms_form" "geography_quiz" {
  title               = "Geography Quiz"
  description         = "Test your geography knowledge!"
  quiz                = true
  published           = true
  accepting_responses = true

  item {
    item_key = "q1_capital"
    multiple_choice {
      question_text = "What is the capital of France?"
      options       = ["Paris", "London", "Rome", "Berlin"]
      required      = true
      grading {
        points             = 5
        correct_answer     = "Paris"
        feedback_correct   = "Correct! Paris has been the capital since the 10th century."
        feedback_incorrect = "Incorrect. The capital of France is Paris."
      }
    }
  }

  item {
    item_key = "q2_river"
    short_answer {
      question_text = "Name the longest river in the world."
      required      = true
      grading {
        points             = 10
        correct_answer     = "Nile"
        feedback_correct   = "Correct!"
        feedback_incorrect = "The answer is the Nile."
      }
    }
  }

  item {
    item_key = "q3_essay"
    paragraph {
      question_text = "Explain why tectonic plates move. (Manually graded)"
      required      = true
      grading {
        points = 20
      }
    }
  }
}

output "quiz_url" {
  value = google_forms_form.geography_quiz.responder_uri
}
