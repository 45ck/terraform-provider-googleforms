// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

// Package convert provides pure functions for translating between
// Terraform state models and Google Forms API request/response structures.
package convert

import (
	"fmt"

	forms "google.golang.org/api/forms/v1"
)

// ItemModelToCreateRequest converts a convert.ItemModel into a Forms API
// Request containing a CreateItemRequest. The index determines the item's
// position in the form.
func ItemModelToCreateRequest(item ItemModel, index int) (*forms.Request, error) {
	formItem := &forms.Item{Title: item.Title}

	switch {
	case item.MultipleChoice != nil:
		buildMultipleChoice(formItem, item.MultipleChoice)
	case item.ShortAnswer != nil:
		buildShortAnswer(formItem, item.ShortAnswer)
	case item.Paragraph != nil:
		buildParagraph(formItem, item.Paragraph)
	default:
		return nil, fmt.Errorf("item %q has no question block set", item.Title)
	}

	return &forms.Request{
		CreateItem: &forms.CreateItemRequest{
			Item:     formItem,
			Location: &forms.Location{Index: int64(index)},
		},
	}, nil
}

// buildMultipleChoice populates a forms.Item with a ChoiceQuestion (RADIO).
func buildMultipleChoice(fi *forms.Item, mc *MultipleChoiceBlock) {
	opts := make([]*forms.Option, len(mc.Options))
	for i, v := range mc.Options {
		opts[i] = &forms.Option{Value: v}
	}

	q := &forms.Question{
		Required: mc.Required,
		ChoiceQuestion: &forms.ChoiceQuestion{
			Type:    "RADIO",
			Options: opts,
		},
	}
	applyGrading(q, mc.Grading)
	fi.QuestionItem = &forms.QuestionItem{Question: q}
}

// buildShortAnswer populates a forms.Item with a TextQuestion (paragraph=false).
func buildShortAnswer(fi *forms.Item, sa *ShortAnswerBlock) {
	q := &forms.Question{
		Required:     sa.Required,
		TextQuestion: &forms.TextQuestion{Paragraph: false},
	}
	applyGrading(q, sa.Grading)
	fi.QuestionItem = &forms.QuestionItem{Question: q}
}

// buildParagraph populates a forms.Item with a TextQuestion (paragraph=true).
func buildParagraph(fi *forms.Item, p *ParagraphBlock) {
	q := &forms.Question{
		Required:     p.Required,
		TextQuestion: &forms.TextQuestion{Paragraph: true},
	}
	applyGrading(q, p.Grading)
	fi.QuestionItem = &forms.QuestionItem{Question: q}
}

// applyGrading sets grading fields on a Question if grading is configured.
func applyGrading(q *forms.Question, g *GradingBlock) {
	if g == nil {
		return
	}
	grading := &forms.Grading{PointValue: g.Points}

	if g.CorrectAnswer != "" {
		grading.CorrectAnswers = &forms.CorrectAnswers{
			Answers: []*forms.CorrectAnswer{{Value: g.CorrectAnswer}},
		}
	}
	if g.FeedbackCorrect != "" {
		grading.WhenRight = &forms.Feedback{Text: g.FeedbackCorrect}
	}
	if g.FeedbackIncorrect != "" {
		grading.WhenWrong = &forms.Feedback{Text: g.FeedbackIncorrect}
	}
	q.Grading = grading
}

// ItemsToCreateRequests converts a slice of ItemModels into an ordered
// slice of Forms API CreateItem Requests.
func ItemsToCreateRequests(items []ItemModel) ([]*forms.Request, error) {
	reqs := make([]*forms.Request, 0, len(items))
	for i, item := range items {
		req, err := ItemModelToCreateRequest(item, i)
		if err != nil {
			return nil, fmt.Errorf("item[%d]: %w", i, err)
		}
		reqs = append(reqs, req)
	}
	return reqs, nil
}

// BuildDeleteRequests creates delete requests for all items in reverse order
// (highest index first) to avoid index shifting during batch deletion.
func BuildDeleteRequests(itemCount int) []*forms.Request {
	reqs := make([]*forms.Request, 0, itemCount)
	for i := itemCount - 1; i >= 0; i-- {
		reqs = append(reqs, &forms.Request{
			DeleteItem: &forms.DeleteItemRequest{
				Location: &forms.Location{Index: int64(i)},
			},
		})
	}
	return reqs
}

// BuildUpdateInfoRequest creates a request to update the form's title and
// description.
func BuildUpdateInfoRequest(title, description string) *forms.Request {
	return &forms.Request{
		UpdateFormInfo: &forms.UpdateFormInfoRequest{
			Info: &forms.Info{
				Title:       title,
				Description: description,
			},
			UpdateMask: "title,description",
		},
	}
}

// BuildQuizSettingsRequest creates a request to enable or disable quiz mode.
func BuildQuizSettingsRequest(isQuiz bool) *forms.Request {
	return &forms.Request{
		UpdateSettings: &forms.UpdateSettingsRequest{
			Settings: &forms.FormSettings{
				QuizSettings: &forms.QuizSettings{
					IsQuiz: isQuiz,
				},
			},
			UpdateMask: "quizSettings.isQuiz",
		},
	}
}
