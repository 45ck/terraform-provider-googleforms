// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package convert

// ItemModel is the convert-package representation of a form item.
// It mirrors resourceform.ItemModel but uses plain Go types to avoid
// circular imports with the Terraform framework types package.
type ItemModel struct {
	Title              string
	ItemKey            string
	GoogleItemID       string
	MultipleChoice     *MultipleChoiceBlock
	ShortAnswer        *ShortAnswerBlock
	Paragraph          *ParagraphBlock
	Dropdown           *DropdownBlock
	Checkbox           *CheckboxBlock
	MultipleChoiceGrid *MultipleChoiceGridBlock
	CheckboxGrid       *CheckboxGridBlock
	Date               *DateBlock
	DateTime           *DateTimeBlock
	Scale              *ScaleBlock
	Time               *TimeBlock
	Rating             *RatingBlock
	FileUpload         *FileUploadBlock
	TextItem           *TextItemBlock
	Image              *ImageBlock
	Video              *VideoBlock
	SectionHeader      *SectionHeaderBlock
}

// MultipleChoiceGridBlock describes a grid question with radio button rows.
type MultipleChoiceGridBlock struct {
	QuestionText     string
	Rows             []string
	Columns          []string
	Required         bool
	ShuffleQuestions bool
	ShuffleColumns   bool
}

// CheckboxGridBlock describes a grid question with checkbox rows.
type CheckboxGridBlock struct {
	QuestionText     string
	Rows             []string
	Columns          []string
	Required         bool
	ShuffleQuestions bool
	ShuffleColumns   bool
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

// DropdownBlock describes a dropdown (select) question.
type DropdownBlock struct {
	QuestionText string
	Options      []string
	Required     bool
	Grading      *GradingBlock
}

// CheckboxBlock describes a checkbox (multi-select) question.
type CheckboxBlock struct {
	QuestionText string
	Options      []string
	Required     bool
	Grading      *GradingBlock
}

// DateBlock describes a date question (no time component).
type DateBlock struct {
	QuestionText string
	Required     bool
	IncludeYear  bool
}

// DateTimeBlock describes a date+time question.
type DateTimeBlock struct {
	QuestionText string
	Required     bool
	IncludeYear  bool
}

// ScaleBlock describes a linear scale question.
type ScaleBlock struct {
	QuestionText string
	Required     bool
	Low          int64
	High         int64
	LowLabel     string
	HighLabel    string
}

// TimeBlock describes a time or duration question.
type TimeBlock struct {
	QuestionText string
	Required     bool
	Duration     bool
}

// RatingBlock describes a rating question.
type RatingBlock struct {
	QuestionText     string
	Required         bool
	IconType         string
	RatingScaleLevel int64
}

// FileUploadBlock describes a file upload question.
//
// Note: The Forms API does not reliably support creating file upload questions.
// This block is primarily for import/management of existing items.
type FileUploadBlock struct {
	QuestionText string
	Required     bool
	FolderID     string
	MaxFileSize  int64
	MaxFiles     int64
	Types        []string
}

// TextItemBlock describes a text-only item (no question).
type TextItemBlock struct {
	Title       string
	Description string
}

// ImageBlock describes an image item (non-question).
type ImageBlock struct {
	Title       string
	Description string
	SourceURI   string
	AltText     string
	ContentURI  string
}

// VideoBlock describes a video item (non-question).
type VideoBlock struct {
	Title       string
	Description string
	YoutubeURI  string
	Caption     string
}

// SectionHeaderBlock describes a section break / page header.
// Unlike question types, this has no QuestionItem - only Title and Description.
type SectionHeaderBlock struct {
	Title       string
	Description string
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
	ID                  string
	Title               string
	Description         string
	DocumentTitle       string
	ResponderURI        string
	RevisionID          string
	Quiz                bool
	EmailCollectionType string
	Items               []ItemModel
}
