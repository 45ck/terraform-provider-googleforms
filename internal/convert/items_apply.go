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

	case desired.MultipleChoiceGrid != nil:
		if existing.QuestionGroupItem == nil || existing.QuestionGroupItem.Grid == nil || existing.QuestionGroupItem.Grid.Columns == nil {
			return true, nil
		}
		if existing.QuestionGroupItem.Grid.Columns.Type != "RADIO" {
			return true, nil
		}
		if len(existing.QuestionGroupItem.Questions) != len(desired.MultipleChoiceGrid.Rows) {
			return true, nil
		}
		if len(existing.QuestionGroupItem.Grid.Columns.Options) != len(desired.MultipleChoiceGrid.Columns) {
			return true, nil
		}

		existing.Title = desired.MultipleChoiceGrid.QuestionText
		existing.QuestionItem = nil
		existing.TextItem = nil
		existing.ImageItem = nil
		existing.VideoItem = nil
		existing.PageBreakItem = nil

		existing.QuestionGroupItem.Grid.ShuffleQuestions = desired.MultipleChoiceGrid.ShuffleQuestions
		existing.QuestionGroupItem.Grid.Columns.Shuffle = desired.MultipleChoiceGrid.ShuffleColumns
		existing.QuestionGroupItem.Grid.Columns.Type = "RADIO"
		for i, v := range desired.MultipleChoiceGrid.Columns {
			existing.QuestionGroupItem.Grid.Columns.Options[i] = &forms.Option{Value: v}
		}

		for i, row := range desired.MultipleChoiceGrid.Rows {
			q := existing.QuestionGroupItem.Questions[i]
			if q == nil {
				q = &forms.Question{}
				existing.QuestionGroupItem.Questions[i] = q
			}
			if q.RowQuestion == nil {
				q.RowQuestion = &forms.RowQuestion{}
			}
			q.RowQuestion.Title = row
			q.Required = desired.MultipleChoiceGrid.Required
			q.ChoiceQuestion = nil
			q.TextQuestion = nil
			q.DateQuestion = nil
			q.ScaleQuestion = nil
			q.TimeQuestion = nil
			q.RatingQuestion = nil
			q.FileUploadQuestion = nil
		}

	case desired.CheckboxGrid != nil:
		if existing.QuestionGroupItem == nil || existing.QuestionGroupItem.Grid == nil || existing.QuestionGroupItem.Grid.Columns == nil {
			return true, nil
		}
		if existing.QuestionGroupItem.Grid.Columns.Type != "CHECKBOX" {
			return true, nil
		}
		if len(existing.QuestionGroupItem.Questions) != len(desired.CheckboxGrid.Rows) {
			return true, nil
		}
		if len(existing.QuestionGroupItem.Grid.Columns.Options) != len(desired.CheckboxGrid.Columns) {
			return true, nil
		}

		existing.Title = desired.CheckboxGrid.QuestionText
		existing.QuestionItem = nil
		existing.TextItem = nil
		existing.ImageItem = nil
		existing.VideoItem = nil
		existing.PageBreakItem = nil

		existing.QuestionGroupItem.Grid.ShuffleQuestions = desired.CheckboxGrid.ShuffleQuestions
		existing.QuestionGroupItem.Grid.Columns.Shuffle = desired.CheckboxGrid.ShuffleColumns
		existing.QuestionGroupItem.Grid.Columns.Type = "CHECKBOX"
		for i, v := range desired.CheckboxGrid.Columns {
			existing.QuestionGroupItem.Grid.Columns.Options[i] = &forms.Option{Value: v}
		}

		for i, row := range desired.CheckboxGrid.Rows {
			q := existing.QuestionGroupItem.Questions[i]
			if q == nil {
				q = &forms.Question{}
				existing.QuestionGroupItem.Questions[i] = q
			}
			if q.RowQuestion == nil {
				q.RowQuestion = &forms.RowQuestion{}
			}
			q.RowQuestion.Title = row
			q.Required = desired.CheckboxGrid.Required
			q.ChoiceQuestion = nil
			q.TextQuestion = nil
			q.DateQuestion = nil
			q.ScaleQuestion = nil
			q.TimeQuestion = nil
			q.RatingQuestion = nil
			q.FileUploadQuestion = nil
		}

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

	case desired.FileUpload != nil:
		// File upload questions are only manageable for existing items. We allow
		// updating title/required but refuse any structural changes.
		if existing.QuestionItem == nil || existing.QuestionItem.Question == nil {
			return true, nil
		}
		q := existing.QuestionItem.Question
		if q.FileUploadQuestion == nil {
			return true, nil
		}

		q.Required = desired.FileUpload.Required
		// Do not mutate file upload settings beyond required flag.
		q.ChoiceQuestion = nil
		q.TextQuestion = nil
		q.DateQuestion = nil
		q.ScaleQuestion = nil
		q.TimeQuestion = nil
		q.RatingQuestion = nil
		q.RowQuestion = nil
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

	case desired.Image != nil:
		if existing.ImageItem == nil || existing.ImageItem.Image == nil {
			return true, nil
		}
		existing.Title = desired.Image.Title
		existing.Description = desired.Image.Description
		// contentUri is output-only; preserve it by only updating inputs.
		existing.ImageItem.Image.SourceUri = desired.Image.SourceURI
		existing.ImageItem.Image.AltText = desired.Image.AltText
		existing.QuestionItem = nil
		existing.QuestionGroupItem = nil
		existing.TextItem = nil
		existing.VideoItem = nil
		existing.PageBreakItem = nil

	case desired.Video != nil:
		if existing.VideoItem == nil || existing.VideoItem.Video == nil {
			return true, nil
		}
		existing.Title = desired.Video.Title
		existing.Description = desired.Video.Description
		existing.VideoItem.Caption = desired.Video.Caption
		existing.VideoItem.Video.YoutubeUri = desired.Video.YoutubeURI
		existing.QuestionItem = nil
		existing.QuestionGroupItem = nil
		existing.TextItem = nil
		existing.ImageItem = nil
		existing.PageBreakItem = nil

	default:
		return true, fmt.Errorf("desired item has no supported question block")
	}

	return false, nil
}
