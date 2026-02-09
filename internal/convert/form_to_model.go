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

	for i, apiItem := range form.Items {
		itemKey := resolveItemKey(apiItem.ItemId, i, existingKeyMap)
		converted, err := FormItemToItemModel(apiItem, itemKey)
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
// Returns nil (without error) for unsupported item types (page breaks, images, etc.).
func FormItemToItemModel(item *forms.Item, itemKey string) (*ItemModel, error) {
	if item.QuestionItem == nil || item.QuestionItem.Question == nil {
		return nil, nil
	}

	q := item.QuestionItem.Question
	model := &ItemModel{
		Title:        item.Title,
		ItemKey:      itemKey,
		GoogleItemID: item.ItemId,
	}

	switch {
	case q.ChoiceQuestion != nil && q.ChoiceQuestion.Type == "RADIO":
		model.MultipleChoice = convertChoiceQuestion(item.Title, q)
	case q.TextQuestion != nil && !q.TextQuestion.Paragraph:
		model.ShortAnswer = convertShortAnswer(item.Title, q)
	case q.TextQuestion != nil && q.TextQuestion.Paragraph:
		model.Paragraph = convertParagraph(item.Title, q)
	default:
		return nil, nil // unsupported question type
	}

	return model, nil
}

// convertChoiceQuestion maps a RADIO ChoiceQuestion to MultipleChoiceBlock.
func convertChoiceQuestion(title string, q *forms.Question) *MultipleChoiceBlock {
	opts := make([]string, len(q.ChoiceQuestion.Options))
	for i, o := range q.ChoiceQuestion.Options {
		opts[i] = o.Value
	}
	mc := &MultipleChoiceBlock{
		QuestionText: title,
		Options:      opts,
		Required:     q.Required,
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
