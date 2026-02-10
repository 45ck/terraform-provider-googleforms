// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

// Package resourceform implements the googleforms_form Terraform resource.
package resourceform

import "github.com/hashicorp/terraform-plugin-framework/types"

// FormResourceModel describes the Terraform state for googleforms_form.
type FormResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Title               types.String `tfsdk:"title"`
	Description         types.String `tfsdk:"description"`
	Published           types.Bool   `tfsdk:"published"`
	AcceptingResponses  types.Bool   `tfsdk:"accepting_responses"`
	Quiz                types.Bool   `tfsdk:"quiz"`
	UpdateStrategy      types.String `tfsdk:"update_strategy"`
	DangerousReplaceAll types.Bool   `tfsdk:"dangerously_replace_all_items"`
	ManageMode          types.String `tfsdk:"manage_mode"`
	ConflictPolicy      types.String `tfsdk:"conflict_policy"`
	Items               types.List   `tfsdk:"item"`
	ContentJSON         types.String `tfsdk:"content_json"`
	ResponderURI        types.String `tfsdk:"responder_uri"`
	EditURI             types.String `tfsdk:"edit_uri"`
	DocumentTitle       types.String `tfsdk:"document_title"`
	RevisionID          types.String `tfsdk:"revision_id"`
}

// ItemModel describes a single form item in Terraform state.
type ItemModel struct {
	ItemKey        types.String         `tfsdk:"item_key"`
	MultipleChoice *MultipleChoiceModel `tfsdk:"multiple_choice"`
	ShortAnswer    *ShortAnswerModel    `tfsdk:"short_answer"`
	Paragraph      *ParagraphModel      `tfsdk:"paragraph"`
	Dropdown       *DropdownModel       `tfsdk:"dropdown"`
	Checkbox       *CheckboxModel       `tfsdk:"checkbox"`
	Date           *DateModel           `tfsdk:"date"`
	DateTime       *DateTimeModel       `tfsdk:"date_time"`
	Scale          *ScaleModel          `tfsdk:"scale"`
	Time           *TimeModel           `tfsdk:"time"`
	Rating         *RatingModel         `tfsdk:"rating"`
	TextItem       *TextItemModel       `tfsdk:"text_item"`
	SectionHeader  *SectionHeaderModel  `tfsdk:"section_header"`
	GoogleItemID   types.String         `tfsdk:"google_item_id"`
}

// MultipleChoiceModel describes a multiple choice question.
type MultipleChoiceModel struct {
	QuestionText types.String  `tfsdk:"question_text"`
	Options      types.List    `tfsdk:"options"`
	Required     types.Bool    `tfsdk:"required"`
	Grading      *GradingModel `tfsdk:"grading"`
}

// ShortAnswerModel describes a short answer question.
type ShortAnswerModel struct {
	QuestionText types.String  `tfsdk:"question_text"`
	Required     types.Bool    `tfsdk:"required"`
	Grading      *GradingModel `tfsdk:"grading"`
}

// ParagraphModel describes a paragraph (long text) question.
type ParagraphModel struct {
	QuestionText types.String  `tfsdk:"question_text"`
	Required     types.Bool    `tfsdk:"required"`
	Grading      *GradingModel `tfsdk:"grading"`
}

// DropdownModel describes a dropdown (select) question.
type DropdownModel struct {
	QuestionText types.String  `tfsdk:"question_text"`
	Options      types.List    `tfsdk:"options"`
	Required     types.Bool    `tfsdk:"required"`
	Grading      *GradingModel `tfsdk:"grading"`
}

// CheckboxModel describes a checkbox (multi-select) question.
type CheckboxModel struct {
	QuestionText types.String  `tfsdk:"question_text"`
	Options      types.List    `tfsdk:"options"`
	Required     types.Bool    `tfsdk:"required"`
	Grading      *GradingModel `tfsdk:"grading"`
}

// DateModel describes a date question (no time component).
type DateModel struct {
	QuestionText types.String `tfsdk:"question_text"`
	Required     types.Bool   `tfsdk:"required"`
	IncludeYear  types.Bool   `tfsdk:"include_year"`
}

// DateTimeModel describes a date+time question.
type DateTimeModel struct {
	QuestionText types.String `tfsdk:"question_text"`
	Required     types.Bool   `tfsdk:"required"`
	IncludeYear  types.Bool   `tfsdk:"include_year"`
}

// ScaleModel describes a linear scale question.
type ScaleModel struct {
	QuestionText types.String `tfsdk:"question_text"`
	Required     types.Bool   `tfsdk:"required"`
	Low          types.Int64  `tfsdk:"low"`
	High         types.Int64  `tfsdk:"high"`
	LowLabel     types.String `tfsdk:"low_label"`
	HighLabel    types.String `tfsdk:"high_label"`
}

// TimeModel describes a time or duration question.
type TimeModel struct {
	QuestionText types.String `tfsdk:"question_text"`
	Required     types.Bool   `tfsdk:"required"`
	Duration     types.Bool   `tfsdk:"duration"`
}

// RatingModel describes a rating question.
type RatingModel struct {
	QuestionText     types.String `tfsdk:"question_text"`
	Required         types.Bool   `tfsdk:"required"`
	IconType         types.String `tfsdk:"icon_type"`
	RatingScaleLevel types.Int64  `tfsdk:"rating_scale_level"`
}

// TextItemModel describes a text-only item (no question).
type TextItemModel struct {
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
}

// SectionHeaderModel describes a section header / page break.
type SectionHeaderModel struct {
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
}

// GradingModel describes quiz grading options for a question.
type GradingModel struct {
	Points            types.Int64  `tfsdk:"points"`
	CorrectAnswer     types.String `tfsdk:"correct_answer"`
	FeedbackCorrect   types.String `tfsdk:"feedback_correct"`
	FeedbackIncorrect types.String `tfsdk:"feedback_incorrect"`
}
