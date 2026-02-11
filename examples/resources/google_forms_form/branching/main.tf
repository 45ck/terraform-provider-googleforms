resource "googleforms_form" "branching" {
  title               = "Branching Example"
  description         = "Shows go_to_section_key navigation using option blocks."
  published           = true
  accepting_responses = true

  item {
    item_key = "start"
    multiple_choice {
      question_text = "Choose a path"
      shuffle       = true
      option {
        value             = "Go to A"
        go_to_section_key = "section_a"
      }
      option {
        value             = "Go to B"
        go_to_section_key = "section_b"
      }
      option {
        value        = "Submit now"
        go_to_action = "SUBMIT_FORM"
      }
    }
  }

  item {
    item_key = "section_a"
    section_header {
      title       = "Section A"
      description = "Questions for path A"
    }
  }

  item {
    item_key = "a_q1"
    short_answer {
      question_text = "A: your name"
      required      = true
    }
  }

  item {
    item_key = "section_b"
    section_header {
      title       = "Section B"
      description = "Questions for path B"
    }
  }

  item {
    item_key = "b_q1"
    paragraph {
      question_text = "B: tell us more"
    }
  }
}

