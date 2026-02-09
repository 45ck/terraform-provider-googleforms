// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package convert

import (
	"testing"

	forms "google.golang.org/api/forms/v1"
)

// ---------------------------------------------------------------------------
// FormToModel
// ---------------------------------------------------------------------------

func TestFormToModel_BasicForm(t *testing.T) {
	form := &forms.Form{
		FormId:       "form-123",
		ResponderUri: "https://docs.google.com/forms/d/e/xxx/viewform",
		RevisionId:   "rev-1",
		Info: &forms.Info{
			Title:         "My Survey",
			DocumentTitle: "My Survey Doc",
		},
	}

	model, err := FormToModel(form, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.ID != "form-123" {
		t.Errorf("ID = %q, want form-123", model.ID)
	}
	if model.Title != "My Survey" {
		t.Errorf("Title = %q, want 'My Survey'", model.Title)
	}
	if model.DocumentTitle != "My Survey Doc" {
		t.Errorf("DocumentTitle = %q, want 'My Survey Doc'", model.DocumentTitle)
	}
	if model.ResponderURI != "https://docs.google.com/forms/d/e/xxx/viewform" {
		t.Errorf("ResponderURI = %q", model.ResponderURI)
	}
	if model.RevisionID != "rev-1" {
		t.Errorf("RevisionID = %q, want rev-1", model.RevisionID)
	}
}

func TestFormToModel_WithDescription(t *testing.T) {
	form := &forms.Form{
		FormId: "form-456",
		Info: &forms.Info{
			Title:       "Survey",
			Description: "Please fill this out.",
		},
	}

	model, err := FormToModel(form, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.Description != "Please fill this out." {
		t.Errorf("Description = %q, want 'Please fill this out.'", model.Description)
	}
}

func TestFormToModel_QuizMode(t *testing.T) {
	form := &forms.Form{
		FormId: "form-quiz",
		Info:   &forms.Info{Title: "Quiz"},
		Settings: &forms.FormSettings{
			QuizSettings: &forms.QuizSettings{
				IsQuiz: true,
			},
		},
	}

	model, err := FormToModel(form, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !model.Quiz {
		t.Error("expected Quiz to be true")
	}
}

func TestFormToModel_WithMultipleChoiceItem(t *testing.T) {
	form := &forms.Form{
		FormId: "form-mc",
		Info:   &forms.Info{Title: "MC Form"},
		Items: []*forms.Item{
			{
				ItemId: "item-abc",
				Title:  "Pick a color",
				QuestionItem: &forms.QuestionItem{
					Question: &forms.Question{
						ChoiceQuestion: &forms.ChoiceQuestion{
							Type: "RADIO",
							Options: []*forms.Option{
								{Value: "Red"},
								{Value: "Blue"},
							},
						},
					},
				},
			},
		},
	}

	keyMap := map[string]string{"item-abc": "color_question"}
	model, err := FormToModel(form, keyMap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(model.Items) != 1 {
		t.Fatalf("items count = %d, want 1", len(model.Items))
	}

	item := model.Items[0]
	if item.ItemKey != "color_question" {
		t.Errorf("ItemKey = %q, want 'color_question'", item.ItemKey)
	}
	if item.GoogleItemID != "item-abc" {
		t.Errorf("GoogleItemID = %q, want 'item-abc'", item.GoogleItemID)
	}
	if item.MultipleChoice == nil {
		t.Fatal("expected MultipleChoice to be set")
	}
	if item.MultipleChoice.QuestionText != "Pick a color" {
		t.Errorf("QuestionText = %q", item.MultipleChoice.QuestionText)
	}
	if len(item.MultipleChoice.Options) != 2 {
		t.Fatalf("options = %d, want 2", len(item.MultipleChoice.Options))
	}
	if item.MultipleChoice.Options[0] != "Red" {
		t.Errorf("option[0] = %q, want Red", item.MultipleChoice.Options[0])
	}
}

func TestFormToModel_WithShortAnswerItem(t *testing.T) {
	form := &forms.Form{
		FormId: "form-sa",
		Info:   &forms.Info{Title: "SA Form"},
		Items: []*forms.Item{
			{
				ItemId: "item-sa1",
				Title:  "Your name?",
				QuestionItem: &forms.QuestionItem{
					Question: &forms.Question{
						Required: true,
						TextQuestion: &forms.TextQuestion{
							Paragraph: false,
						},
					},
				},
			},
		},
	}

	model, err := FormToModel(form, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(model.Items) != 1 {
		t.Fatalf("items count = %d, want 1", len(model.Items))
	}

	item := model.Items[0]
	if item.ShortAnswer == nil {
		t.Fatal("expected ShortAnswer to be set")
	}
	if item.ShortAnswer.QuestionText != "Your name?" {
		t.Errorf("QuestionText = %q", item.ShortAnswer.QuestionText)
	}
	if !item.ShortAnswer.Required {
		t.Error("expected Required to be true")
	}
}

func TestFormToModel_WithParagraphItem(t *testing.T) {
	form := &forms.Form{
		FormId: "form-para",
		Info:   &forms.Info{Title: "Para Form"},
		Items: []*forms.Item{
			{
				ItemId: "item-p1",
				Title:  "Describe yourself",
				QuestionItem: &forms.QuestionItem{
					Question: &forms.Question{
						TextQuestion: &forms.TextQuestion{
							Paragraph: true,
						},
					},
				},
			},
		},
	}

	model, err := FormToModel(form, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(model.Items) != 1 {
		t.Fatalf("items count = %d, want 1", len(model.Items))
	}

	item := model.Items[0]
	if item.Paragraph == nil {
		t.Fatal("expected Paragraph to be set")
	}
	if item.Paragraph.QuestionText != "Describe yourself" {
		t.Errorf("QuestionText = %q", item.Paragraph.QuestionText)
	}
}

func TestFormToModel_WithGrading(t *testing.T) {
	form := &forms.Form{
		FormId: "form-graded",
		Info:   &forms.Info{Title: "Graded"},
		Items: []*forms.Item{
			{
				ItemId: "item-g1",
				Title:  "Capital of France?",
				QuestionItem: &forms.QuestionItem{
					Question: &forms.Question{
						ChoiceQuestion: &forms.ChoiceQuestion{
							Type: "RADIO",
							Options: []*forms.Option{
								{Value: "London"},
								{Value: "Paris"},
							},
						},
						Grading: &forms.Grading{
							PointValue: 5,
							CorrectAnswers: &forms.CorrectAnswers{
								Answers: []*forms.CorrectAnswer{
									{Value: "Paris"},
								},
							},
							WhenRight: &forms.Feedback{Text: "Correct!"},
							WhenWrong: &forms.Feedback{Text: "Wrong!"},
						},
					},
				},
			},
		},
	}

	model, err := FormToModel(form, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(model.Items) != 1 {
		t.Fatalf("items count = %d, want 1", len(model.Items))
	}

	mc := model.Items[0].MultipleChoice
	if mc == nil {
		t.Fatal("expected MultipleChoice")
	}
	g := mc.Grading
	if g == nil {
		t.Fatal("expected Grading to be set")
	}
	if g.Points != 5 {
		t.Errorf("Points = %d, want 5", g.Points)
	}
	if g.CorrectAnswer != "Paris" {
		t.Errorf("CorrectAnswer = %q, want Paris", g.CorrectAnswer)
	}
	if g.FeedbackCorrect != "Correct!" {
		t.Errorf("FeedbackCorrect = %q", g.FeedbackCorrect)
	}
	if g.FeedbackIncorrect != "Wrong!" {
		t.Errorf("FeedbackIncorrect = %q", g.FeedbackIncorrect)
	}
}

func TestFormToModel_MultipleItems_PreservesOrder(t *testing.T) {
	form := &forms.Form{
		FormId: "form-multi",
		Info:   &forms.Info{Title: "Multi"},
		Items: []*forms.Item{
			{
				ItemId: "id-1",
				Title:  "First",
				QuestionItem: &forms.QuestionItem{
					Question: &forms.Question{
						TextQuestion: &forms.TextQuestion{Paragraph: false},
					},
				},
			},
			{
				ItemId: "id-2",
				Title:  "Second",
				QuestionItem: &forms.QuestionItem{
					Question: &forms.Question{
						TextQuestion: &forms.TextQuestion{Paragraph: true},
					},
				},
			},
			{
				ItemId: "id-3",
				Title:  "Third",
				QuestionItem: &forms.QuestionItem{
					Question: &forms.Question{
						ChoiceQuestion: &forms.ChoiceQuestion{
							Type:    "RADIO",
							Options: []*forms.Option{{Value: "A"}},
						},
					},
				},
			},
		},
	}

	model, err := FormToModel(form, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(model.Items) != 3 {
		t.Fatalf("items count = %d, want 3", len(model.Items))
	}

	// Verify order preserved
	if model.Items[0].ShortAnswer == nil {
		t.Error("item[0]: expected ShortAnswer")
	}
	if model.Items[1].Paragraph == nil {
		t.Error("item[1]: expected Paragraph")
	}
	if model.Items[2].MultipleChoice == nil {
		t.Error("item[2]: expected MultipleChoice")
	}
}

func TestFormToModel_EmptyForm(t *testing.T) {
	form := &forms.Form{
		FormId: "form-empty",
		Info:   &forms.Info{Title: "Empty"},
	}

	model, err := FormToModel(form, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(model.Items) != 0 {
		t.Errorf("items count = %d, want 0", len(model.Items))
	}
}

func TestFormToModel_UnsupportedItemType_SkipsGracefully(t *testing.T) {
	form := &forms.Form{
		FormId: "form-unsupported",
		Info:   &forms.Info{Title: "Has Unsupported"},
		Items: []*forms.Item{
			{
				ItemId: "id-supported",
				Title:  "Supported",
				QuestionItem: &forms.QuestionItem{
					Question: &forms.Question{
						TextQuestion: &forms.TextQuestion{Paragraph: false},
					},
				},
			},
			{
				// An item with no QuestionItem (e.g. PageBreak, Image, Video)
				ItemId: "id-unsupported",
				Title:  "Page Break",
			},
		},
	}

	model, err := FormToModel(form, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only the supported item should be included
	if len(model.Items) != 1 {
		t.Fatalf("items count = %d, want 1", len(model.Items))
	}
	if model.Items[0].ShortAnswer == nil {
		t.Error("expected ShortAnswer for the supported item")
	}
}

// ---------------------------------------------------------------------------
// FormItemToItemModel
// ---------------------------------------------------------------------------

func TestFormItemToItemModel_NilQuestionItem(t *testing.T) {
	item := &forms.Item{
		ItemId: "no-question",
		Title:  "Section Header",
	}

	result, err := FormItemToItemModel(item, "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil result for unsupported item type")
	}
}
