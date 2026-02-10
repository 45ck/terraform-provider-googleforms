// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package convert

import (
	"fmt"

	forms "google.golang.org/api/forms/v1"
)

// ApplyItemModelToExistingItem updates an existing Forms API Item in-place to
// match the desired ItemModel, attempting to preserve item/question IDs.
//
// It returns (true, nil) when the change is structural (e.g. question type
// changed) and cannot be applied safely as an in-place update.
func ApplyItemModelToExistingItem(existing *forms.Item, desired ItemModel) (bool, error) {
	if existing == nil {
		return true, fmt.Errorf("existing item is nil")
	}

	// Always update title/description where applicable.
	// For question items, Title is the question text.
	existing.Title = desired.Title

	switch {
	case desired.MultipleChoice != nil:
		if existing.QuestionItem == nil || existing.QuestionItem.Question == nil {
			return true, nil
		}
		q := existing.QuestionItem.Question
		// Refuse to "retarget" a different question type in-place.
		if q.ChoiceQuestion == nil || q.ChoiceQuestion.Type != "RADIO" {
			return true, nil
		}

		q.Required = desired.MultipleChoice.Required
		q.TextQuestion = nil
		q.DateQuestion = nil
		q.ScaleQuestion = nil
		q.TimeQuestion = nil
		q.RatingQuestion = nil
		q.FileUploadQuestion = nil
		q.RowQuestion = nil

		opts := make([]*forms.Option, len(desired.MultipleChoice.Options))
		for i, v := range desired.MultipleChoice.Options {
			opts[i] = &forms.Option{Value: v}
		}
		q.ChoiceQuestion.Options = opts
		applyGrading(q, desired.MultipleChoice.Grading)

		existing.PageBreakItem = nil
		existing.TextItem = nil
		existing.ImageItem = nil
		existing.VideoItem = nil
		existing.QuestionGroupItem = nil

	case desired.Dropdown != nil:
		if existing.QuestionItem == nil || existing.QuestionItem.Question == nil {
			return true, nil
		}
		q := existing.QuestionItem.Question
		if q.ChoiceQuestion == nil || q.ChoiceQuestion.Type != "DROP_DOWN" {
			return true, nil
		}

		q.Required = desired.Dropdown.Required
		q.TextQuestion = nil
		q.DateQuestion = nil
		q.ScaleQuestion = nil
		q.TimeQuestion = nil
		q.RatingQuestion = nil
		q.FileUploadQuestion = nil
		q.RowQuestion = nil

		opts := make([]*forms.Option, len(desired.Dropdown.Options))
		for i, v := range desired.Dropdown.Options {
			opts[i] = &forms.Option{Value: v}
		}
		q.ChoiceQuestion.Options = opts
		applyGrading(q, desired.Dropdown.Grading)

		existing.PageBreakItem = nil
		existing.TextItem = nil
		existing.ImageItem = nil
		existing.VideoItem = nil
		existing.QuestionGroupItem = nil

	case desired.Checkbox != nil:
		if existing.QuestionItem == nil || existing.QuestionItem.Question == nil {
			return true, nil
		}
		q := existing.QuestionItem.Question
		if q.ChoiceQuestion == nil || q.ChoiceQuestion.Type != "CHECKBOX" {
			return true, nil
		}

		q.Required = desired.Checkbox.Required
		q.TextQuestion = nil
		q.DateQuestion = nil
		q.ScaleQuestion = nil
		q.TimeQuestion = nil
		q.RatingQuestion = nil
		q.FileUploadQuestion = nil
		q.RowQuestion = nil

		opts := make([]*forms.Option, len(desired.Checkbox.Options))
		for i, v := range desired.Checkbox.Options {
			opts[i] = &forms.Option{Value: v}
		}
		q.ChoiceQuestion.Options = opts
		applyGrading(q, desired.Checkbox.Grading)

		existing.PageBreakItem = nil
		existing.TextItem = nil
		existing.ImageItem = nil
		existing.VideoItem = nil
		existing.QuestionGroupItem = nil

	case desired.ShortAnswer != nil:
		if existing.QuestionItem == nil || existing.QuestionItem.Question == nil {
			return true, nil
		}
		q := existing.QuestionItem.Question
		if q.TextQuestion == nil || q.TextQuestion.Paragraph {
			return true, nil
		}

		q.Required = desired.ShortAnswer.Required
		q.ChoiceQuestion = nil
		q.DateQuestion = nil
		q.ScaleQuestion = nil
		q.TimeQuestion = nil
		q.RatingQuestion = nil
		q.FileUploadQuestion = nil
		q.RowQuestion = nil

		q.TextQuestion.Paragraph = false
		applyGrading(q, desired.ShortAnswer.Grading)

		existing.PageBreakItem = nil
		existing.TextItem = nil
		existing.ImageItem = nil
		existing.VideoItem = nil
		existing.QuestionGroupItem = nil

	case desired.Paragraph != nil:
		if existing.QuestionItem == nil || existing.QuestionItem.Question == nil {
			return true, nil
		}
		q := existing.QuestionItem.Question
		if q.TextQuestion == nil || !q.TextQuestion.Paragraph {
			return true, nil
		}

		q.Required = desired.Paragraph.Required
		q.ChoiceQuestion = nil
		q.DateQuestion = nil
		q.ScaleQuestion = nil
		q.TimeQuestion = nil
		q.RatingQuestion = nil
		q.FileUploadQuestion = nil
		q.RowQuestion = nil

		q.TextQuestion.Paragraph = true
		applyGrading(q, desired.Paragraph.Grading)

		existing.PageBreakItem = nil
		existing.TextItem = nil
		existing.ImageItem = nil
		existing.VideoItem = nil
		existing.QuestionGroupItem = nil

	case desired.Date != nil:
		if existing.QuestionItem == nil || existing.QuestionItem.Question == nil {
			return true, nil
		}
		q := existing.QuestionItem.Question
		if q.DateQuestion == nil || q.DateQuestion.IncludeTime {
			return true, nil
		}

		q.Required = desired.Date.Required
		q.ChoiceQuestion = nil
		q.TextQuestion = nil
		q.ScaleQuestion = nil
		q.TimeQuestion = nil
		q.RatingQuestion = nil
		q.FileUploadQuestion = nil
		q.RowQuestion = nil

		q.DateQuestion.IncludeTime = false
		q.DateQuestion.IncludeYear = desired.Date.IncludeYear
		q.Grading = nil // date questions cannot be graded in current schema

		existing.PageBreakItem = nil
		existing.TextItem = nil
		existing.ImageItem = nil
		existing.VideoItem = nil
		existing.QuestionGroupItem = nil

	case desired.DateTime != nil:
		if existing.QuestionItem == nil || existing.QuestionItem.Question == nil {
			return true, nil
		}
		q := existing.QuestionItem.Question
		if q.DateQuestion == nil || !q.DateQuestion.IncludeTime {
			return true, nil
		}

		q.Required = desired.DateTime.Required
		q.ChoiceQuestion = nil
		q.TextQuestion = nil
		q.ScaleQuestion = nil
		q.TimeQuestion = nil
		q.RatingQuestion = nil
		q.FileUploadQuestion = nil
		q.RowQuestion = nil

		q.DateQuestion.IncludeTime = true
		q.DateQuestion.IncludeYear = desired.DateTime.IncludeYear
		q.Grading = nil

		existing.PageBreakItem = nil
		existing.TextItem = nil
		existing.ImageItem = nil
		existing.VideoItem = nil
		existing.QuestionGroupItem = nil

	case desired.Scale != nil:
		if existing.QuestionItem == nil || existing.QuestionItem.Question == nil {
			return true, nil
		}
		q := existing.QuestionItem.Question
		if q.ScaleQuestion == nil {
			return true, nil
		}

		q.Required = desired.Scale.Required
		q.ChoiceQuestion = nil
		q.TextQuestion = nil
		q.DateQuestion = nil
		q.TimeQuestion = nil
		q.RatingQuestion = nil
		q.FileUploadQuestion = nil
		q.RowQuestion = nil

		q.ScaleQuestion.Low = desired.Scale.Low
		q.ScaleQuestion.High = desired.Scale.High
		q.ScaleQuestion.LowLabel = desired.Scale.LowLabel
		q.ScaleQuestion.HighLabel = desired.Scale.HighLabel
		q.Grading = nil

		existing.PageBreakItem = nil
		existing.TextItem = nil
		existing.ImageItem = nil
		existing.VideoItem = nil
		existing.QuestionGroupItem = nil

	case desired.SectionHeader != nil:
		// Section header is represented as a page break with title/description.
		if existing.PageBreakItem == nil {
			return true, nil
		}
		existing.Title = desired.SectionHeader.Title
		existing.Description = desired.SectionHeader.Description
		existing.QuestionItem = nil
		existing.QuestionGroupItem = nil
		existing.TextItem = nil
		existing.ImageItem = nil
		existing.VideoItem = nil

	default:
		return true, fmt.Errorf("desired item has no supported question block")
	}

	return false, nil
}
