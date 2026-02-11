// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// 5. OptionsRequiredForChoiceValidator
// ---------------------------------------------------------------------------

// OptionsRequiredForChoiceValidator ensures multiple_choice questions have
// at least one option.
type OptionsRequiredForChoiceValidator struct{}

func (v OptionsRequiredForChoiceValidator) Description(_ context.Context) string {
	return "Validates that choice/grid questions have required options/rows."
}

func (v OptionsRequiredForChoiceValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v OptionsRequiredForChoiceValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var items types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("item"), &items)...)

	if resp.Diagnostics.HasError() || items.IsNull() || items.IsUnknown() {
		return
	}

	var itemModels []ItemModel
	resp.Diagnostics.Append(items.ElementsAs(ctx, &itemModels, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	for _, item := range itemModels {
		if item.MultipleChoice != nil {
			opts := item.MultipleChoice.Options
			if opts.IsNull() || opts.IsUnknown() || len(opts.Elements()) == 0 {
				resp.Diagnostics.AddError(
					"Missing Options",
					"A multiple_choice question requires at least one option.",
				)
				return
			}
		}
		if item.Dropdown != nil {
			opts := item.Dropdown.Options
			if opts.IsNull() || opts.IsUnknown() || len(opts.Elements()) == 0 {
				resp.Diagnostics.AddError(
					"Missing Options",
					"A dropdown question requires at least one option.",
				)
				return
			}
		}
		if item.Checkbox != nil {
			opts := item.Checkbox.Options
			if opts.IsNull() || opts.IsUnknown() || len(opts.Elements()) == 0 {
				resp.Diagnostics.AddError(
					"Missing Options",
					"A checkbox question requires at least one option.",
				)
				return
			}
		}
		if item.MultipleChoiceGrid != nil {
			if item.MultipleChoiceGrid.Rows.IsNull() || item.MultipleChoiceGrid.Rows.IsUnknown() || len(item.MultipleChoiceGrid.Rows.Elements()) == 0 {
				resp.Diagnostics.AddError(
					"Missing Rows",
					"A multiple_choice_grid question requires at least one row.",
				)
				return
			}
			if item.MultipleChoiceGrid.Columns.IsNull() || item.MultipleChoiceGrid.Columns.IsUnknown() || len(item.MultipleChoiceGrid.Columns.Elements()) == 0 {
				resp.Diagnostics.AddError(
					"Missing Columns",
					"A multiple_choice_grid question requires at least one column.",
				)
				return
			}
		}
		if item.CheckboxGrid != nil {
			if item.CheckboxGrid.Rows.IsNull() || item.CheckboxGrid.Rows.IsUnknown() || len(item.CheckboxGrid.Rows.Elements()) == 0 {
				resp.Diagnostics.AddError(
					"Missing Rows",
					"A checkbox_grid question requires at least one row.",
				)
				return
			}
			if item.CheckboxGrid.Columns.IsNull() || item.CheckboxGrid.Columns.IsUnknown() || len(item.CheckboxGrid.Columns.Elements()) == 0 {
				resp.Diagnostics.AddError(
					"Missing Columns",
					"A checkbox_grid question requires at least one column.",
				)
				return
			}
		}
	}
}

// ---------------------------------------------------------------------------
// 6. CorrectAnswerInOptionsValidator
// ---------------------------------------------------------------------------

// CorrectAnswerInOptionsValidator ensures that when a correct_answer is
// specified for a multiple_choice question, it matches one of the options.
type CorrectAnswerInOptionsValidator struct{}

func (v CorrectAnswerInOptionsValidator) Description(_ context.Context) string {
	return "Validates that correct_answer matches an option for multiple_choice."
}

func (v CorrectAnswerInOptionsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v CorrectAnswerInOptionsValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var items types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("item"), &items)...)

	if resp.Diagnostics.HasError() || items.IsNull() || items.IsUnknown() {
		return
	}

	var itemModels []ItemModel
	resp.Diagnostics.Append(items.ElementsAs(ctx, &itemModels, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	for _, item := range itemModels {
		if item.MultipleChoice != nil && item.MultipleChoice.Grading != nil {
			checkCorrectAnswerInChoiceOptions(ctx, item.MultipleChoice.Grading, item.MultipleChoice.Options, resp)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		if item.Dropdown != nil && item.Dropdown.Grading != nil {
			checkCorrectAnswerInChoiceOptions(ctx, item.Dropdown.Grading, item.Dropdown.Options, resp)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}
}

// checkCorrectAnswerInChoiceOptions validates that the correct_answer is in the options list.
func checkCorrectAnswerInChoiceOptions(
	ctx context.Context,
	grading *GradingModel,
	optionsList types.List,
	resp *resource.ValidateConfigResponse,
) {
	answer := grading.CorrectAnswer
	if answer.IsNull() || answer.IsUnknown() {
		return
	}

	answerVal := answer.ValueString()

	var options []types.String
	resp.Diagnostics.Append(optionsList.ElementsAs(ctx, &options, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	for _, opt := range options {
		if opt.ValueString() == answerVal {
			return
		}
	}

	resp.Diagnostics.AddError(
		"Invalid Correct Answer",
		`The correct_answer "`+answerVal+`" is not in the options list.`,
	)
}

// ---------------------------------------------------------------------------
// 7. GradingRequiresQuizValidator
// ---------------------------------------------------------------------------

// GradingRequiresQuizValidator ensures that grading blocks are only used when
// quiz mode is enabled on the form.
type GradingRequiresQuizValidator struct{}

func (v GradingRequiresQuizValidator) Description(_ context.Context) string {
	return "Validates that grading blocks require quiz = true."
}

func (v GradingRequiresQuizValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v GradingRequiresQuizValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var quiz types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("quiz"), &quiz)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If quiz is true (or null/unknown), no error is possible.
	if quiz.IsNull() || quiz.IsUnknown() || quiz.ValueBool() {
		return
	}

	if hasAnyGrading(ctx, req, resp) {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Grading requires quiz mode. Add quiz = true to the form resource.",
		)
	}
}

// hasAnyGrading returns true if any item has a grading block set.
func hasAnyGrading(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) bool {
	var items types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("item"), &items)...)

	if resp.Diagnostics.HasError() || items.IsNull() || items.IsUnknown() {
		return false
	}

	var itemModels []ItemModel
	resp.Diagnostics.Append(items.ElementsAs(ctx, &itemModels, false)...)

	if resp.Diagnostics.HasError() {
		return false
	}

	for _, item := range itemModels {
		if itemHasGrading(item) {
			return true
		}
	}
	return false
}

// itemHasGrading returns true if the item has a grading sub-block.
func itemHasGrading(item ItemModel) bool {
	if item.MultipleChoice != nil && item.MultipleChoice.Grading != nil {
		return true
	}
	if item.ShortAnswer != nil && item.ShortAnswer.Grading != nil {
		return true
	}
	if item.Paragraph != nil && item.Paragraph.Grading != nil {
		return true
	}
	if item.Dropdown != nil && item.Dropdown.Grading != nil {
		return true
	}
	if item.Checkbox != nil && item.Checkbox.Grading != nil {
		return true
	}
	return false
}
