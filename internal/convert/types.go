// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package convert

// ItemModel is the convert-package representation of a form item.
// It mirrors resourceform.ItemModel but uses plain Go types to avoid
// circular imports with the Terraform framework types package.
type ItemModel struct {
	Title          string
	ItemKey        string
	GoogleItemID   string
	MultipleChoice *MultipleChoiceBlock
	ShortAnswer    *ShortAnswerBlock
	Paragraph      *ParagraphBlock
}

// MultipleChoiceBlock describes a multiple-choice (radio) question.
type MultipleChoiceBlock struct {
	QuestionText string
	Options      []string
	Required     bool
	Grading      *GradingBlock
}

// ShortAnswerBlock describes a short-answer (single-line text) question.
type ShortAnswerBlock struct {
	QuestionText string
	Required     bool
	Grading      *GradingBlock
}

// ParagraphBlock describes a paragraph (multi-line text) question.
type ParagraphBlock struct {
	QuestionText string
	Required     bool
	Grading      *GradingBlock
}

// GradingBlock describes quiz grading settings for a question.
type GradingBlock struct {
	Points            int64
	CorrectAnswer     string
	FeedbackCorrect   string
	FeedbackIncorrect string
}

// FormModel is the convert-package representation of the full form state.
// Uses plain Go types instead of Terraform framework types.
type FormModel struct {
	ID            string
	Title         string
	Description   string
	DocumentTitle string
	ResponderURI  string
	RevisionID    string
	Quiz          bool
	Items         []ItemModel
}
