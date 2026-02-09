// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/45ck/terraform-provider-googleforms/internal/convert"
)

// tfItemsToConvertItems extracts items from the Terraform plan's types.List
// and converts them to []convert.ItemModel for use by the convert package.
func tfItemsToConvertItems(ctx context.Context, items types.List) ([]convert.ItemModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if items.IsNull() || items.IsUnknown() || len(items.Elements()) == 0 {
		return nil, diags
	}

	var tfItems []ItemModel
	diags.Append(items.ElementsAs(ctx, &tfItems, false)...)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]convert.ItemModel, len(tfItems))
	for i, tf := range tfItems {
		result[i] = convert.ItemModel{
			ItemKey:      tf.ItemKey.ValueString(),
			GoogleItemID: tf.GoogleItemID.ValueString(),
		}

		if tf.MultipleChoice != nil {
			mc := &convert.MultipleChoiceBlock{
				QuestionText: tf.MultipleChoice.QuestionText.ValueString(),
				Required:     tf.MultipleChoice.Required.ValueBool(),
			}

			var opts []string
			diags.Append(tf.MultipleChoice.Options.ElementsAs(ctx, &opts, false)...)
			if diags.HasError() {
				return nil, diags
			}
			mc.Options = opts

			if tf.MultipleChoice.Grading != nil {
				mc.Grading = tfGradingToConvert(tf.MultipleChoice.Grading)
			}

			result[i].MultipleChoice = mc
			result[i].Title = mc.QuestionText
		}

		if tf.ShortAnswer != nil {
			sa := &convert.ShortAnswerBlock{
				QuestionText: tf.ShortAnswer.QuestionText.ValueString(),
				Required:     tf.ShortAnswer.Required.ValueBool(),
			}
			if tf.ShortAnswer.Grading != nil {
				sa.Grading = tfGradingToConvert(tf.ShortAnswer.Grading)
			}
			result[i].ShortAnswer = sa
			result[i].Title = sa.QuestionText
		}

		if tf.Paragraph != nil {
			p := &convert.ParagraphBlock{
				QuestionText: tf.Paragraph.QuestionText.ValueString(),
				Required:     tf.Paragraph.Required.ValueBool(),
			}
			if tf.Paragraph.Grading != nil {
				p.Grading = tfGradingToConvert(tf.Paragraph.Grading)
			}
			result[i].Paragraph = p
			result[i].Title = p.QuestionText
		}

		if tf.Dropdown != nil {
			dd := &convert.DropdownBlock{
				QuestionText: tf.Dropdown.QuestionText.ValueString(),
				Required:     tf.Dropdown.Required.ValueBool(),
			}
			var opts []string
			diags.Append(tf.Dropdown.Options.ElementsAs(ctx, &opts, false)...)
			if diags.HasError() {
				return nil, diags
			}
			dd.Options = opts
			if tf.Dropdown.Grading != nil {
				dd.Grading = tfGradingToConvert(tf.Dropdown.Grading)
			}
			result[i].Dropdown = dd
			result[i].Title = dd.QuestionText
		}

		if tf.Checkbox != nil {
			cb := &convert.CheckboxBlock{
				QuestionText: tf.Checkbox.QuestionText.ValueString(),
				Required:     tf.Checkbox.Required.ValueBool(),
			}
			var opts []string
			diags.Append(tf.Checkbox.Options.ElementsAs(ctx, &opts, false)...)
			if diags.HasError() {
				return nil, diags
			}
			cb.Options = opts
			if tf.Checkbox.Grading != nil {
				cb.Grading = tfGradingToConvert(tf.Checkbox.Grading)
			}
			result[i].Checkbox = cb
			result[i].Title = cb.QuestionText
		}

		if tf.Date != nil {
			result[i].Date = &convert.DateBlock{
				QuestionText: tf.Date.QuestionText.ValueString(),
				Required:     tf.Date.Required.ValueBool(),
				IncludeYear:  tf.Date.IncludeYear.ValueBool(),
			}
			result[i].Title = tf.Date.QuestionText.ValueString()
		}

		if tf.DateTime != nil {
			result[i].DateTime = &convert.DateTimeBlock{
				QuestionText: tf.DateTime.QuestionText.ValueString(),
				Required:     tf.DateTime.Required.ValueBool(),
				IncludeYear:  tf.DateTime.IncludeYear.ValueBool(),
			}
			result[i].Title = tf.DateTime.QuestionText.ValueString()
		}

		if tf.Scale != nil {
			result[i].Scale = &convert.ScaleBlock{
				QuestionText: tf.Scale.QuestionText.ValueString(),
				Required:     tf.Scale.Required.ValueBool(),
				Low:          tf.Scale.Low.ValueInt64(),
				High:         tf.Scale.High.ValueInt64(),
				LowLabel:     tf.Scale.LowLabel.ValueString(),
				HighLabel:    tf.Scale.HighLabel.ValueString(),
			}
			result[i].Title = tf.Scale.QuestionText.ValueString()
		}

		if tf.SectionHeader != nil {
			result[i].SectionHeader = &convert.SectionHeaderBlock{
				Title:       tf.SectionHeader.Title.ValueString(),
				Description: tf.SectionHeader.Description.ValueString(),
			}
			result[i].Title = tf.SectionHeader.Title.ValueString()
		}
	}

	return result, diags
}

// tfGradingToConvert converts a TF GradingModel to a convert.GradingBlock.
func tfGradingToConvert(g *GradingModel) *convert.GradingBlock {
	return &convert.GradingBlock{
		Points:            g.Points.ValueInt64(),
		CorrectAnswer:     g.CorrectAnswer.ValueString(),
		FeedbackCorrect:   g.FeedbackCorrect.ValueString(),
		FeedbackIncorrect: g.FeedbackIncorrect.ValueString(),
	}
}

// convertFormModelToTFState maps a convert.FormModel back into a
// FormResourceModel, preserving plan values for non-computed fields.
func convertFormModelToTFState(model *convert.FormModel, plan FormResourceModel) FormResourceModel {
	state := FormResourceModel{
		ID:            types.StringValue(model.ID),
		Title:         types.StringValue(model.Title),
		Description:   types.StringValue(model.Description),
		Published:     plan.Published,
		AcceptingResponses: plan.AcceptingResponses,
		Quiz:          types.BoolValue(model.Quiz),
		ContentJSON:   plan.ContentJSON,
		ResponderURI:  types.StringValue(model.ResponderURI),
		DocumentTitle: types.StringValue(model.DocumentTitle),
	}

	// edit_uri follows a known pattern
	if model.ID != "" {
		state.EditURI = types.StringValue("https://docs.google.com/forms/d/" + model.ID + "/edit")
	}

	return state
}

// buildItemKeyMap creates a mapping from google_item_id -> item_key from
// current state items, enabling correlation of API items back to TF config.
func buildItemKeyMap(ctx context.Context, items types.List) (map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics

	if items.IsNull() || items.IsUnknown() || len(items.Elements()) == 0 {
		return nil, diags
	}

	var tfItems []ItemModel
	diags.Append(items.ElementsAs(ctx, &tfItems, false)...)
	if diags.HasError() {
		return nil, diags
	}

	keyMap := make(map[string]string, len(tfItems))
	for _, item := range tfItems {
		googleID := item.GoogleItemID.ValueString()
		itemKey := item.ItemKey.ValueString()
		if googleID != "" && itemKey != "" {
			keyMap[googleID] = itemKey
		}
	}

	return keyMap, diags
}

// convertItemsToTFList converts []convert.ItemModel back to types.List
// for setting in Terraform state.
func convertItemsToTFList(ctx context.Context, items []convert.ItemModel) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(items) == 0 {
		return types.ListNull(itemObjectType()), diags
	}

	tfItems := make([]ItemModel, len(items))
	for i, item := range items {
		tfItems[i] = convertItemModelToTF(ctx, item, &diags)
		if diags.HasError() {
			return types.ListNull(itemObjectType()), diags
		}
	}

	result, d := types.ListValueFrom(ctx, itemObjectType(), tfItems)
	diags.Append(d...)
	return result, diags
}

// convertItemModelToTF converts a single convert.ItemModel to the TF ItemModel.
func convertItemModelToTF(ctx context.Context, item convert.ItemModel, diags *diag.Diagnostics) ItemModel {
	tf := ItemModel{
		ItemKey:      types.StringValue(item.ItemKey),
		GoogleItemID: types.StringValue(item.GoogleItemID),
	}

	if item.MultipleChoice != nil {
		mc := item.MultipleChoice
		opts, d := types.ListValueFrom(ctx, types.StringType, mc.Options)
		diags.Append(d...)

		tf.MultipleChoice = &MultipleChoiceModel{
			QuestionText: types.StringValue(mc.QuestionText),
			Options:      opts,
			Required:     types.BoolValue(mc.Required),
		}
		if mc.Grading != nil {
			tf.MultipleChoice.Grading = convertGradingToTF(mc.Grading)
		}
	}

	if item.ShortAnswer != nil {
		sa := item.ShortAnswer
		tf.ShortAnswer = &ShortAnswerModel{
			QuestionText: types.StringValue(sa.QuestionText),
			Required:     types.BoolValue(sa.Required),
		}
		if sa.Grading != nil {
			tf.ShortAnswer.Grading = convertGradingToTF(sa.Grading)
		}
	}

	if item.Paragraph != nil {
		p := item.Paragraph
		tf.Paragraph = &ParagraphModel{
			QuestionText: types.StringValue(p.QuestionText),
			Required:     types.BoolValue(p.Required),
		}
		if p.Grading != nil {
			tf.Paragraph.Grading = convertGradingToTF(p.Grading)
		}
	}

	if item.Dropdown != nil {
		dd := item.Dropdown
		opts, d := types.ListValueFrom(ctx, types.StringType, dd.Options)
		diags.Append(d...)
		tf.Dropdown = &DropdownModel{
			QuestionText: types.StringValue(dd.QuestionText),
			Options:      opts,
			Required:     types.BoolValue(dd.Required),
		}
		if dd.Grading != nil {
			tf.Dropdown.Grading = convertGradingToTF(dd.Grading)
		}
	}

	if item.Checkbox != nil {
		cb := item.Checkbox
		opts, d := types.ListValueFrom(ctx, types.StringType, cb.Options)
		diags.Append(d...)
		tf.Checkbox = &CheckboxModel{
			QuestionText: types.StringValue(cb.QuestionText),
			Options:      opts,
			Required:     types.BoolValue(cb.Required),
		}
		if cb.Grading != nil {
			tf.Checkbox.Grading = convertGradingToTF(cb.Grading)
		}
	}

	if item.Date != nil {
		tf.Date = &DateModel{
			QuestionText: types.StringValue(item.Date.QuestionText),
			Required:     types.BoolValue(item.Date.Required),
			IncludeYear:  types.BoolValue(item.Date.IncludeYear),
		}
	}

	if item.DateTime != nil {
		tf.DateTime = &DateTimeModel{
			QuestionText: types.StringValue(item.DateTime.QuestionText),
			Required:     types.BoolValue(item.DateTime.Required),
			IncludeYear:  types.BoolValue(item.DateTime.IncludeYear),
		}
	}

	if item.Scale != nil {
		s := item.Scale
		tf.Scale = &ScaleModel{
			QuestionText: types.StringValue(s.QuestionText),
			Required:     types.BoolValue(s.Required),
			Low:          types.Int64Value(s.Low),
			High:         types.Int64Value(s.High),
		}
		if s.LowLabel != "" {
			tf.Scale.LowLabel = types.StringValue(s.LowLabel)
		} else {
			tf.Scale.LowLabel = types.StringNull()
		}
		if s.HighLabel != "" {
			tf.Scale.HighLabel = types.StringValue(s.HighLabel)
		} else {
			tf.Scale.HighLabel = types.StringNull()
		}
	}

	if item.SectionHeader != nil {
		sh := item.SectionHeader
		tf.SectionHeader = &SectionHeaderModel{
			Title: types.StringValue(sh.Title),
		}
		if sh.Description != "" {
			tf.SectionHeader.Description = types.StringValue(sh.Description)
		} else {
			tf.SectionHeader.Description = types.StringNull()
		}
	}

	return tf
}

// convertGradingToTF converts a convert.GradingBlock to a TF GradingModel.
func convertGradingToTF(g *convert.GradingBlock) *GradingModel {
	gm := &GradingModel{
		Points: types.Int64Value(g.Points),
	}

	if g.CorrectAnswer != "" {
		gm.CorrectAnswer = types.StringValue(g.CorrectAnswer)
	} else {
		gm.CorrectAnswer = types.StringNull()
	}

	if g.FeedbackCorrect != "" {
		gm.FeedbackCorrect = types.StringValue(g.FeedbackCorrect)
	} else {
		gm.FeedbackCorrect = types.StringNull()
	}

	if g.FeedbackIncorrect != "" {
		gm.FeedbackIncorrect = types.StringValue(g.FeedbackIncorrect)
	} else {
		gm.FeedbackIncorrect = types.StringNull()
	}

	return gm
}

// itemObjectType returns the types.ObjectType for a single item block element.
// This must match the schema defined in schema.go.
func itemObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"item_key":       types.StringType,
			"google_item_id": types.StringType,
			"multiple_choice": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text": types.StringType,
					"options":       types.ListType{ElemType: types.StringType},
					"required":      types.BoolType,
					"grading":       gradingObjectType(),
				},
			},
			"short_answer": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text": types.StringType,
					"required":      types.BoolType,
					"grading":       gradingObjectType(),
				},
			},
			"paragraph": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text": types.StringType,
					"required":      types.BoolType,
					"grading":       gradingObjectType(),
				},
			},
			"dropdown": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text": types.StringType,
					"options":       types.ListType{ElemType: types.StringType},
					"required":      types.BoolType,
					"grading":       gradingObjectType(),
				},
			},
			"checkbox": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text": types.StringType,
					"options":       types.ListType{ElemType: types.StringType},
					"required":      types.BoolType,
					"grading":       gradingObjectType(),
				},
			},
			"date": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text": types.StringType,
					"required":      types.BoolType,
					"include_year":  types.BoolType,
				},
			},
			"date_time": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text": types.StringType,
					"required":      types.BoolType,
					"include_year":  types.BoolType,
				},
			},
			"scale": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"question_text": types.StringType,
					"required":      types.BoolType,
					"low":           types.Int64Type,
					"high":          types.Int64Type,
					"low_label":     types.StringType,
					"high_label":    types.StringType,
				},
			},
			"section_header": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"title":       types.StringType,
					"description": types.StringType,
				},
			},
		},
	}
}

// gradingObjectType returns the types.ObjectType for the grading block.
func gradingObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"points":             types.Int64Type,
			"correct_answer":     types.StringType,
			"feedback_correct":   types.StringType,
			"feedback_incorrect": types.StringType,
		},
	}
}
