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
	case item.Dropdown != nil:
		buildDropdown(formItem, item.Dropdown)
	case item.Checkbox != nil:
		buildCheckbox(formItem, item.Checkbox)
	case item.MultipleChoiceGrid != nil:
		buildMultipleChoiceGrid(formItem, item.MultipleChoiceGrid)
	case item.CheckboxGrid != nil:
		buildCheckboxGrid(formItem, item.CheckboxGrid)
	case item.Date != nil:
		buildDate(formItem, item.Date)
	case item.DateTime != nil:
		buildDateTime(formItem, item.DateTime)
	case item.Scale != nil:
		buildScale(formItem, item.Scale)
	case item.Time != nil:
		buildTime(formItem, item.Time)
	case item.Rating != nil:
		buildRating(formItem, item.Rating)
	case item.FileUpload != nil:
		return nil, fmt.Errorf("file upload questions cannot be created via the Forms API; import the form or use content_json/forms_batch_update")
	case item.TextItem != nil:
		buildTextItem(formItem, item.TextItem)
	case item.Image != nil:
		buildImage(formItem, item.Image)
	case item.Video != nil:
		buildVideo(formItem, item.Video)
	case item.SectionHeader != nil:
		buildSectionHeader(formItem, item.SectionHeader)
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

func buildMultipleChoiceGrid(fi *forms.Item, g *MultipleChoiceGridBlock) {
	fi.Title = g.QuestionText
	opts := make([]*forms.Option, len(g.Columns))
	for i, v := range g.Columns {
		opts[i] = &forms.Option{Value: v}
	}

	qs := make([]*forms.Question, len(g.Rows))
	for i, row := range g.Rows {
		qs[i] = &forms.Question{
			Required:    g.Required,
			RowQuestion: &forms.RowQuestion{Title: row},
		}
	}

	fi.QuestionGroupItem = &forms.QuestionGroupItem{
		Grid: &forms.Grid{
			Columns: &forms.ChoiceQuestion{
				Type:    "RADIO",
				Options: opts,
				Shuffle: g.ShuffleColumns,
			},
			ShuffleQuestions: g.ShuffleQuestions,
		},
		Questions: qs,
	}
	fi.QuestionItem = nil
	fi.TextItem = nil
	fi.PageBreakItem = nil
	fi.ImageItem = nil
	fi.VideoItem = nil
}

func buildCheckboxGrid(fi *forms.Item, g *CheckboxGridBlock) {
	fi.Title = g.QuestionText
	opts := make([]*forms.Option, len(g.Columns))
	for i, v := range g.Columns {
		opts[i] = &forms.Option{Value: v}
	}

	qs := make([]*forms.Question, len(g.Rows))
	for i, row := range g.Rows {
		qs[i] = &forms.Question{
			Required:    g.Required,
			RowQuestion: &forms.RowQuestion{Title: row},
		}
	}

	fi.QuestionGroupItem = &forms.QuestionGroupItem{
		Grid: &forms.Grid{
			Columns: &forms.ChoiceQuestion{
				Type:    "CHECKBOX",
				Options: opts,
				Shuffle: g.ShuffleColumns,
			},
			ShuffleQuestions: g.ShuffleQuestions,
		},
		Questions: qs,
	}
	fi.QuestionItem = nil
	fi.TextItem = nil
	fi.PageBreakItem = nil
	fi.ImageItem = nil
	fi.VideoItem = nil
}

// buildMultipleChoice populates a forms.Item with a ChoiceQuestion (RADIO).
func buildMultipleChoice(fi *forms.Item, mc *MultipleChoiceBlock) {
	opts := choiceOptionsToAPI(mc.Options, mc.HasOther)

	q := &forms.Question{
		Required: mc.Required,
		ChoiceQuestion: &forms.ChoiceQuestion{
			Type:    "RADIO",
			Options: opts,
			Shuffle: mc.Shuffle,
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

// buildDropdown populates a forms.Item with a ChoiceQuestion (DROP_DOWN).
func buildDropdown(fi *forms.Item, dd *DropdownBlock) {
	opts := choiceOptionsToAPI(dd.Options, false)

	q := &forms.Question{
		Required: dd.Required,
		ChoiceQuestion: &forms.ChoiceQuestion{
			Type:    "DROP_DOWN",
			Options: opts,
			Shuffle: dd.Shuffle,
		},
	}
	applyGrading(q, dd.Grading)
	fi.QuestionItem = &forms.QuestionItem{Question: q}
}

// buildCheckbox populates a forms.Item with a ChoiceQuestion (CHECKBOX).
func buildCheckbox(fi *forms.Item, cb *CheckboxBlock) {
	opts := choiceOptionsToAPI(cb.Options, cb.HasOther)

	q := &forms.Question{
		Required: cb.Required,
		ChoiceQuestion: &forms.ChoiceQuestion{
			Type:    "CHECKBOX",
			Options: opts,
			Shuffle: cb.Shuffle,
		},
	}
	applyGrading(q, cb.Grading)
	fi.QuestionItem = &forms.QuestionItem{Question: q}
}

func choiceOptionsToAPI(opts []ChoiceOption, includeOther bool) []*forms.Option {
	out := make([]*forms.Option, 0, len(opts)+1)
	for _, o := range opts {
		apiOpt := &forms.Option{Value: o.Value}
		if o.GoToAction != "" && o.GoToAction != "GO_TO_ACTION_UNSPECIFIED" {
			apiOpt.GoToAction = o.GoToAction
		}
		if o.GoToSectionID != "" {
			apiOpt.GoToSectionId = o.GoToSectionID
		}
		out = append(out, apiOpt)
	}
	if includeOther {
		// Note: Forms API uses IsOther to represent the "Other" option; Value is still required.
		out = append(out, &forms.Option{Value: "Other", IsOther: true})
	}
	return out
}

// buildDate populates a forms.Item with a DateQuestion (no time).
func buildDate(fi *forms.Item, d *DateBlock) {
	q := &forms.Question{
		Required: d.Required,
		DateQuestion: &forms.DateQuestion{
			IncludeTime: false,
			IncludeYear: d.IncludeYear,
		},
	}
	fi.QuestionItem = &forms.QuestionItem{Question: q}
}

// buildDateTime populates a forms.Item with a DateQuestion (with time).
func buildDateTime(fi *forms.Item, dt *DateTimeBlock) {
	q := &forms.Question{
		Required: dt.Required,
		DateQuestion: &forms.DateQuestion{
			IncludeTime: true,
			IncludeYear: dt.IncludeYear,
		},
	}
	fi.QuestionItem = &forms.QuestionItem{Question: q}
}

// buildScale populates a forms.Item with a ScaleQuestion.
func buildScale(fi *forms.Item, s *ScaleBlock) {
	q := &forms.Question{
		Required: s.Required,
		ScaleQuestion: &forms.ScaleQuestion{
			Low:       s.Low,
			High:      s.High,
			LowLabel:  s.LowLabel,
			HighLabel: s.HighLabel,
		},
	}
	fi.QuestionItem = &forms.QuestionItem{Question: q}
}

// buildTime populates a forms.Item with a TimeQuestion.
func buildTime(fi *forms.Item, t *TimeBlock) {
	q := &forms.Question{
		Required:     t.Required,
		TimeQuestion: &forms.TimeQuestion{Duration: t.Duration},
	}
	fi.QuestionItem = &forms.QuestionItem{Question: q}
}

// buildRating populates a forms.Item with a RatingQuestion.
func buildRating(fi *forms.Item, r *RatingBlock) {
	q := &forms.Question{
		Required: r.Required,
		RatingQuestion: &forms.RatingQuestion{
			IconType:         r.IconType,
			RatingScaleLevel: r.RatingScaleLevel,
		},
	}
	fi.QuestionItem = &forms.QuestionItem{Question: q}
}

// buildTextItem populates a forms.Item as a text-only item.
func buildTextItem(fi *forms.Item, t *TextItemBlock) {
	fi.Title = t.Title
	fi.Description = t.Description
	fi.TextItem = &forms.TextItem{}
	fi.QuestionItem = nil
	fi.PageBreakItem = nil
}

// buildImage populates a forms.Item as an image item.
func buildImage(fi *forms.Item, img *ImageBlock) {
	fi.Title = img.Title
	fi.Description = img.Description
	fi.ImageItem = &forms.ImageItem{
		Image: &forms.Image{
			SourceUri: img.SourceURI,
			AltText:   img.AltText,
		},
	}
	fi.QuestionItem = nil
	fi.TextItem = nil
	fi.PageBreakItem = nil
	fi.VideoItem = nil
	fi.QuestionGroupItem = nil
}

// buildVideo populates a forms.Item as a video item.
func buildVideo(fi *forms.Item, v *VideoBlock) {
	fi.Title = v.Title
	fi.Description = v.Description
	fi.VideoItem = &forms.VideoItem{
		Caption: v.Caption,
		Video: &forms.Video{
			YoutubeUri: v.YoutubeURI,
		},
	}
	fi.QuestionItem = nil
	fi.TextItem = nil
	fi.PageBreakItem = nil
	fi.ImageItem = nil
	fi.QuestionGroupItem = nil
}

// buildSectionHeader populates a forms.Item as a section header.
// Section headers have Title and Description but no QuestionItem.
func buildSectionHeader(fi *forms.Item, sh *SectionHeaderBlock) {
	fi.Title = sh.Title
	fi.Description = sh.Description
	// No QuestionItem - this creates a page break / section header.
	fi.PageBreakItem = &forms.PageBreakItem{}
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

// BuildEmailCollectionTypeRequest sets the email collection type on the form.
func BuildEmailCollectionTypeRequest(emailCollectionType string) *forms.Request {
	return &forms.Request{
		UpdateSettings: &forms.UpdateSettingsRequest{
			Settings: &forms.FormSettings{
				EmailCollectionType: emailCollectionType,
			},
			UpdateMask: "emailCollectionType",
		},
	}
}
