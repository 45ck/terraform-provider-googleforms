// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package convert

import (
	"encoding/json"
	"fmt"

	forms "google.golang.org/api/forms/v1"
)

// ApplyDesiredItem returns an updated copy of the existing item matching the
// desired model. It preserves item/question IDs by starting from a copy of the
// existing item. If the desired change is structural (e.g. question type
// changed), it returns needsReplace=true.
func ApplyDesiredItem(existing *forms.Item, desired ItemModel) (updated *forms.Item, changed bool, needsReplace bool, err error) {
	if existing == nil {
		return nil, false, true, fmt.Errorf("existing item is nil")
	}

	// Deep copy existing to avoid mutating objects reused elsewhere.
	var copy forms.Item
	b, merr := json.Marshal(existing)
	if merr != nil {
		return nil, false, true, fmt.Errorf("marshal existing item: %w", merr)
	}
	if uerr := json.Unmarshal(b, &copy); uerr != nil {
		return nil, false, true, fmt.Errorf("unmarshal existing item: %w", uerr)
	}

	// Apply changes to the copy.
	needsReplace, err = ApplyItemModelToExistingItem(&copy, desired)
	if err != nil || needsReplace {
		return nil, false, needsReplace, err
	}

	// Compare serialized forms as a pragmatic change detector.
	after, aerr := json.Marshal(&copy)
	if aerr != nil {
		return nil, false, true, fmt.Errorf("marshal updated item: %w", aerr)
	}
	changed = string(b) != string(after)
	return &copy, changed, false, nil
}

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
		if desired.MultipleChoice.Grading == nil {
			q.Grading = nil
		} else {
			applyGrading(q, desired.MultipleChoice.Grading)
		}

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
		if desired.Dropdown.Grading == nil {
			q.Grading = nil
		} else {
			applyGrading(q, desired.Dropdown.Grading)
		}

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
		if desired.Checkbox.Grading == nil {
			q.Grading = nil
		} else {
			applyGrading(q, desired.Checkbox.Grading)
		}

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
		if desired.ShortAnswer.Grading == nil {
			q.Grading = nil
		} else {
			applyGrading(q, desired.ShortAnswer.Grading)
		}

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
		if desired.Paragraph.Grading == nil {
			q.Grading = nil
		} else {
			applyGrading(q, desired.Paragraph.Grading)
		}

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

	case desired.Time != nil:
		if existing.QuestionItem == nil || existing.QuestionItem.Question == nil {
			return true, nil
		}
		q := existing.QuestionItem.Question
		if q.TimeQuestion == nil {
			return true, nil
		}
		q.Required = desired.Time.Required
		q.ChoiceQuestion = nil
		q.TextQuestion = nil
		q.DateQuestion = nil
		q.ScaleQuestion = nil
		q.RatingQuestion = nil
		q.FileUploadQuestion = nil
		q.RowQuestion = nil
		q.TimeQuestion.Duration = desired.Time.Duration
		q.Grading = nil
		existing.PageBreakItem = nil
		existing.TextItem = nil
		existing.ImageItem = nil
		existing.VideoItem = nil
		existing.QuestionGroupItem = nil

	case desired.Rating != nil:
		if existing.QuestionItem == nil || existing.QuestionItem.Question == nil {
			return true, nil
		}
		q := existing.QuestionItem.Question
		if q.RatingQuestion == nil {
			return true, nil
		}
		q.Required = desired.Rating.Required
		q.ChoiceQuestion = nil
		q.TextQuestion = nil
		q.DateQuestion = nil
		q.ScaleQuestion = nil
		q.TimeQuestion = nil
		q.FileUploadQuestion = nil
		q.RowQuestion = nil
		q.RatingQuestion.IconType = desired.Rating.IconType
		q.RatingQuestion.RatingScaleLevel = desired.Rating.RatingScaleLevel
		q.Grading = nil
		existing.PageBreakItem = nil
		existing.TextItem = nil
		existing.ImageItem = nil
		existing.VideoItem = nil
		existing.QuestionGroupItem = nil

	case desired.TextItem != nil:
		if existing.TextItem == nil {
			return true, nil
		}
		existing.Title = desired.TextItem.Title
		existing.Description = desired.TextItem.Description
		existing.QuestionItem = nil
		existing.QuestionGroupItem = nil
		existing.PageBreakItem = nil
		existing.ImageItem = nil
		existing.VideoItem = nil

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
