// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package convert

import (
	"fmt"

	forms "google.golang.org/api/forms/v1"
)

// FormToModel converts a Forms API response into a convert.FormModel.
// existingKeyMap maps Google item IDs to Terraform item_key values; if nil
// or missing an entry, keys are auto-generated as "item_N".
func FormToModel(form *forms.Form, existingKeyMap map[string]string) (*FormModel, error) {
	model := &FormModel{
		ID:           form.FormId,
		ResponderURI: form.ResponderUri,
		RevisionID:   form.RevisionId,
	}

	if form.Info != nil {
		model.Title = form.Info.Title
		model.Description = form.Info.Description
		model.DocumentTitle = form.Info.DocumentTitle
	}

	if form.Settings != nil && form.Settings.QuizSettings != nil {
		model.Quiz = form.Settings.QuizSettings.IsQuiz
	}
	if form.Settings != nil && form.Settings.EmailCollectionType != "" {
		model.EmailCollectionType = form.Settings.EmailCollectionType
	}

	for i, apiItem := range form.Items {
		itemKey := resolveItemKey(apiItem.ItemId, i, existingKeyMap)
		converted, err := FormItemToItemModel(apiItem, itemKey, existingKeyMap)
		if err != nil {
			return nil, fmt.Errorf("item[%d] (%s): %w", i, apiItem.ItemId, err)
		}
		if converted == nil {
			continue // unsupported item type, skip
		}
		model.Items = append(model.Items, *converted)
	}

	return model, nil
}

// resolveItemKey looks up the Terraform item_key for a Google item ID.
// Falls back to "item_N" when no mapping exists (e.g. during import).
func resolveItemKey(googleID string, index int, keyMap map[string]string) string {
	if keyMap != nil {
		if key, ok := keyMap[googleID]; ok {
			return key
		}
	}
	return fmt.Sprintf("item_%d", index)
}

// FormItemToItemModel converts a single Forms API Item into a convert.ItemModel.
// Returns nil (without error) for truly unsupported item types (images, videos).
func FormItemToItemModel(item *forms.Item, itemKey string, keyMap map[string]string) (*ItemModel, error) {
	model := &ItemModel{
		Title:        item.Title,
		ItemKey:      itemKey,
		GoogleItemID: item.ItemId,
	}

	if item.TextItem != nil {
		model.TextItem = &TextItemBlock{
			Title:       item.Title,
			Description: item.Description,
		}
		return model, nil
	}

	if item.ImageItem != nil && item.ImageItem.Image != nil {
		model.Image = &ImageBlock{
			Title:       item.Title,
			Description: item.Description,
			SourceURI:   item.ImageItem.Image.SourceUri,
			AltText:     item.ImageItem.Image.AltText,
			ContentURI:  item.ImageItem.Image.ContentUri,
		}
		return model, nil
	}

	if item.VideoItem != nil && item.VideoItem.Video != nil {
		model.Video = &VideoBlock{
			Title:       item.Title,
			Description: item.Description,
			YoutubeURI:  item.VideoItem.Video.YoutubeUri,
			Caption:     item.VideoItem.Caption,
		}
		return model, nil
	}

	// Section headers / page breaks have no QuestionItem.
	if item.PageBreakItem != nil {
		model.SectionHeader = &SectionHeaderBlock{
			Title:       item.Title,
			Description: item.Description,
		}
		return model, nil
	}

	if item.QuestionItem == nil || item.QuestionItem.Question == nil {
		// QuestionGroupItem is supported separately below.
		if item.QuestionGroupItem == nil {
			return nil, nil // truly unsupported
		}
	}

	// QuestionGroupItem: grid questions.
	if item.QuestionGroupItem != nil && item.QuestionGroupItem.Grid != nil && item.QuestionGroupItem.Grid.Columns != nil {
		cols := item.QuestionGroupItem.Grid.Columns
		colVals := make([]string, 0, len(cols.Options))
		for _, o := range cols.Options {
			if o != nil {
				colVals = append(colVals, o.Value)
			}
		}
		rowVals := make([]string, 0, len(item.QuestionGroupItem.Questions))
		required := false
		for i, q := range item.QuestionGroupItem.Questions {
			if q == nil || q.RowQuestion == nil {
				continue
			}
			rowVals = append(rowVals, q.RowQuestion.Title)
			if i == 0 {
				required = q.Required
			}
		}

		switch cols.Type {
		case "RADIO":
			model.MultipleChoiceGrid = &MultipleChoiceGridBlock{
				QuestionText:     item.Title,
				Rows:             rowVals,
				Columns:          colVals,
				Required:         required,
				ShuffleQuestions: item.QuestionGroupItem.Grid.ShuffleQuestions,
				ShuffleColumns:   cols.Shuffle,
			}
		case "CHECKBOX":
			model.CheckboxGrid = &CheckboxGridBlock{
				QuestionText:     item.Title,
				Rows:             rowVals,
				Columns:          colVals,
				Required:         required,
				ShuffleQuestions: item.QuestionGroupItem.Grid.ShuffleQuestions,
				ShuffleColumns:   cols.Shuffle,
			}
		default:
			return nil, nil
		}

		return model, nil
	}

	if item.QuestionItem == nil || item.QuestionItem.Question == nil {
		return nil, nil
	}

	q := item.QuestionItem.Question

	switch {
	case q.ChoiceQuestion != nil && q.ChoiceQuestion.Type == "RADIO":
		model.MultipleChoice = convertChoiceQuestion(item.Title, q, keyMap)
	case q.ChoiceQuestion != nil && q.ChoiceQuestion.Type == "DROP_DOWN":
		model.Dropdown = convertDropdownQuestion(item.Title, q, keyMap)
	case q.ChoiceQuestion != nil && q.ChoiceQuestion.Type == "CHECKBOX":
		model.Checkbox = convertCheckboxQuestion(item.Title, q, keyMap)
	case q.TextQuestion != nil && !q.TextQuestion.Paragraph:
		model.ShortAnswer = convertShortAnswer(item.Title, q)
	case q.TextQuestion != nil && q.TextQuestion.Paragraph:
		model.Paragraph = convertParagraph(item.Title, q)
	case q.DateQuestion != nil && !q.DateQuestion.IncludeTime:
		model.Date = convertDateQuestion(item.Title, q)
	case q.DateQuestion != nil && q.DateQuestion.IncludeTime:
		model.DateTime = convertDateTimeQuestion(item.Title, q)
	case q.ScaleQuestion != nil:
		model.Scale = convertScaleQuestion(item.Title, q)
	case q.TimeQuestion != nil:
		model.Time = &TimeBlock{QuestionText: item.Title, Required: q.Required, Duration: q.TimeQuestion.Duration}
	case q.RatingQuestion != nil:
		model.Rating = &RatingBlock{
			QuestionText:     item.Title,
			Required:         q.Required,
			IconType:         q.RatingQuestion.IconType,
			RatingScaleLevel: q.RatingQuestion.RatingScaleLevel,
		}
	case q.FileUploadQuestion != nil:
		fu := q.FileUploadQuestion
		model.FileUpload = &FileUploadBlock{
			QuestionText: item.Title,
			Required:     q.Required,
			FolderID:     fu.FolderId,
			MaxFileSize:  fu.MaxFileSize,
			MaxFiles:     fu.MaxFiles,
			Types:        append([]string{}, fu.Types...),
		}
	default:
		return nil, nil // unsupported question type
	}

	return model, nil
}

// convertChoiceQuestion maps a RADIO ChoiceQuestion to MultipleChoiceBlock.
func convertChoiceQuestion(title string, q *forms.Question, keyMap map[string]string) *MultipleChoiceBlock {
	opts, hasOther := convertChoiceOptions(q.ChoiceQuestion.Options, keyMap)
	mc := &MultipleChoiceBlock{
		QuestionText: title,
		Options:      opts,
		Required:     q.Required,
		Shuffle:      q.ChoiceQuestion.Shuffle,
		HasOther:     hasOther,
	}
	mc.Grading = convertGrading(q.Grading)
	return mc
}

// convertShortAnswer maps a non-paragraph TextQuestion to ShortAnswerBlock.
func convertShortAnswer(title string, q *forms.Question) *ShortAnswerBlock {
	sa := &ShortAnswerBlock{
		QuestionText: title,
		Required:     q.Required,
	}
	sa.Grading = convertGrading(q.Grading)
	return sa
}

// convertParagraph maps a paragraph TextQuestion to ParagraphBlock.
func convertParagraph(title string, q *forms.Question) *ParagraphBlock {
	p := &ParagraphBlock{
		QuestionText: title,
		Required:     q.Required,
	}
	p.Grading = convertGrading(q.Grading)
	return p
}

// convertDropdownQuestion maps a DROP_DOWN ChoiceQuestion to DropdownBlock.
func convertDropdownQuestion(title string, q *forms.Question, keyMap map[string]string) *DropdownBlock {
	opts, _ := convertChoiceOptions(q.ChoiceQuestion.Options, keyMap)
	dd := &DropdownBlock{
		QuestionText: title,
		Options:      opts,
		Required:     q.Required,
		Shuffle:      q.ChoiceQuestion.Shuffle,
	}
	dd.Grading = convertGrading(q.Grading)
	return dd
}

// convertCheckboxQuestion maps a CHECKBOX ChoiceQuestion to CheckboxBlock.
func convertCheckboxQuestion(title string, q *forms.Question, keyMap map[string]string) *CheckboxBlock {
	opts, hasOther := convertChoiceOptions(q.ChoiceQuestion.Options, keyMap)
	cb := &CheckboxBlock{
		QuestionText: title,
		Options:      opts,
		Required:     q.Required,
		Shuffle:      q.ChoiceQuestion.Shuffle,
		HasOther:     hasOther,
	}
	cb.Grading = convertGrading(q.Grading)
	return cb
}

func convertChoiceOptions(apiOpts []*forms.Option, keyMap map[string]string) ([]ChoiceOption, bool) {
	opts := make([]ChoiceOption, 0, len(apiOpts))
	hasOther := false
	for _, o := range apiOpts {
		if o == nil {
			continue
		}
		if o.IsOther {
			hasOther = true
			continue
		}
		co := ChoiceOption{
			Value:         o.Value,
			GoToAction:    o.GoToAction,
			GoToSectionID: o.GoToSectionId,
		}
		if o.GoToSectionId != "" && keyMap != nil {
			if k, ok := keyMap[o.GoToSectionId]; ok {
				co.GoToSectionKey = k
			}
		}
		opts = append(opts, co)
	}
	return opts, hasOther
}

// convertDateQuestion maps a DateQuestion (no time) to DateBlock.
func convertDateQuestion(title string, q *forms.Question) *DateBlock {
	return &DateBlock{
		QuestionText: title,
		Required:     q.Required,
		IncludeYear:  q.DateQuestion.IncludeYear,
	}
}

// convertDateTimeQuestion maps a DateQuestion (with time) to DateTimeBlock.
func convertDateTimeQuestion(title string, q *forms.Question) *DateTimeBlock {
	return &DateTimeBlock{
		QuestionText: title,
		Required:     q.Required,
		IncludeYear:  q.DateQuestion.IncludeYear,
	}
}

// convertScaleQuestion maps a ScaleQuestion to ScaleBlock.
func convertScaleQuestion(title string, q *forms.Question) *ScaleBlock {
	return &ScaleBlock{
		QuestionText: title,
		Required:     q.Required,
		Low:          q.ScaleQuestion.Low,
		High:         q.ScaleQuestion.High,
		LowLabel:     q.ScaleQuestion.LowLabel,
		HighLabel:    q.ScaleQuestion.HighLabel,
	}
}

// convertGrading maps Forms API Grading to a GradingBlock, or nil if absent.
func convertGrading(g *forms.Grading) *GradingBlock {
	if g == nil {
		return nil
	}
	gb := &GradingBlock{Points: g.PointValue}

	if g.CorrectAnswers != nil && len(g.CorrectAnswers.Answers) > 0 {
		gb.CorrectAnswer = g.CorrectAnswers.Answers[0].Value
	}
	if g.WhenRight != nil {
		gb.FeedbackCorrect = g.WhenRight.Text
	}
	if g.WhenWrong != nil {
		gb.FeedbackIncorrect = g.WhenWrong.Text
	}
	return gb
}
