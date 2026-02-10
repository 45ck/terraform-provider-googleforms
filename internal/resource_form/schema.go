// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the Terraform schema for googleforms_form.
func (r *FormResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages a Google Form.",
		Attributes:  formAttributes(),
		Blocks:      formBlocks(),
	}
}

// formAttributes returns the top-level attribute definitions.
func formAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The Google Form ID.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"title": schema.StringAttribute{
			Required:    true,
			Description: "The form title displayed to respondents.",
		},
		"description": schema.StringAttribute{
			Optional:    true,
			Description: "The form description.",
		},
		"published": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
			Description: "Whether the form is published. Must be true before accepting_responses can be true.",
		},
		"accepting_responses": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
			Description: "Whether the form is accepting responses. Requires published = true.",
		},
		"quiz": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
			Description: "Enable quiz mode with grading.",
		},
		"update_strategy": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("replace_all"),
			Description: "Update strategy for form items. 'replace_all' deletes and recreates all items on changes. 'targeted' applies deletes/moves/updates/creates using batchUpdate when item_keys are already correlated to google_item_id in state; it refuses question type changes and does not support content_json.",
			Validators: []validator.String{
				stringvalidator.OneOf("replace_all", "targeted"),
			},
		},
		"dangerously_replace_all_items": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
			Description: "Acknowledge that replace_all item updates can break response mappings and integrations. When false, the provider will emit warnings when replace_all is used.",
		},
		"content_json": schema.StringAttribute{
			Optional:    true,
			Description: "Declarative JSON array of form items. Mutually exclusive with item blocks. Use jsonencode().",
			PlanModifiers: []planmodifier.String{
				ContentJSONHashModifier{},
			},
		},
		"responder_uri": schema.StringAttribute{
			Computed:    true,
			Description: "The URL for respondents to fill out the form.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"edit_uri": schema.StringAttribute{
			Computed:    true,
			Description: "The URL to edit the form.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"document_title": schema.StringAttribute{
			Computed:    true,
			Description: "The Google Drive document title.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

// formBlocks returns the block definitions (item list).
func formBlocks() map[string]schema.Block {
	return map[string]schema.Block{
		"item": schema.ListNestedBlock{
			Description: "A form item (question). Each item requires a unique item_key and exactly one question type sub-block.",
			NestedObject: schema.NestedBlockObject{
				Attributes: itemAttributes(),
				Blocks:     itemBlocks(),
			},
		},
	}
}

// itemAttributes returns the attribute definitions for a single item.
func itemAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"item_key": schema.StringAttribute{
			Required:    true,
			Description: "Unique identifier for this item within the form. Used for stable state tracking. Format: [a-z][a-z0-9_]{0,63}.",
		},
		"google_item_id": schema.StringAttribute{
			Computed:    true,
			Description: "The Google-assigned item ID.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

// itemBlocks returns the nested block definitions for question types.
func itemBlocks() map[string]schema.Block {
	gradingBlock := gradingBlockSchema()
	return map[string]schema.Block{
		"multiple_choice": schema.SingleNestedBlock{
			Description: "A multiple choice (radio button) question.",
			Attributes: map[string]schema.Attribute{
				"question_text": schema.StringAttribute{
					Required:    true,
					Description: "The question text.",
				},
				"options": schema.ListAttribute{
					Required:    true,
					ElementType: types.StringType,
					Description: "List of answer options. Must have at least one.",
				},
				"required": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(false),
					Description: "Whether the question is required.",
				},
			},
			Blocks: map[string]schema.Block{
				"grading": gradingBlock,
			},
		},
		"short_answer": schema.SingleNestedBlock{
			Description: "A short answer (single-line text) question.",
			Attributes: map[string]schema.Attribute{
				"question_text": schema.StringAttribute{
					Required:    true,
					Description: "The question text.",
				},
				"required": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(false),
					Description: "Whether the question is required.",
				},
			},
			Blocks: map[string]schema.Block{
				"grading": gradingBlock,
			},
		},
		"paragraph": schema.SingleNestedBlock{
			Description: "A paragraph (multi-line text) question.",
			Attributes: map[string]schema.Attribute{
				"question_text": schema.StringAttribute{
					Required:    true,
					Description: "The question text.",
				},
				"required": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(false),
					Description: "Whether the question is required.",
				},
			},
			Blocks: map[string]schema.Block{
				"grading": gradingBlock,
			},
		},
		"dropdown": schema.SingleNestedBlock{
			Description: "A dropdown (select) question.",
			Attributes: map[string]schema.Attribute{
				"question_text": schema.StringAttribute{
					Required:    true,
					Description: "The question text.",
				},
				"options": schema.ListAttribute{
					Required:    true,
					ElementType: types.StringType,
					Description: "List of answer options. Must have at least one.",
				},
				"required": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(false),
					Description: "Whether the question is required.",
				},
			},
			Blocks: map[string]schema.Block{
				"grading": gradingBlock,
			},
		},
		"checkbox": schema.SingleNestedBlock{
			Description: "A checkbox (multi-select) question.",
			Attributes: map[string]schema.Attribute{
				"question_text": schema.StringAttribute{
					Required:    true,
					Description: "The question text.",
				},
				"options": schema.ListAttribute{
					Required:    true,
					ElementType: types.StringType,
					Description: "List of answer options. Must have at least one.",
				},
				"required": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(false),
					Description: "Whether the question is required.",
				},
			},
			Blocks: map[string]schema.Block{
				"grading": gradingBlock,
			},
		},
		"date": schema.SingleNestedBlock{
			Description: "A date question (no time component).",
			Attributes: map[string]schema.Attribute{
				"question_text": schema.StringAttribute{
					Required:    true,
					Description: "The question text.",
				},
				"required": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(false),
					Description: "Whether the question is required.",
				},
				"include_year": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(true),
					Description: "Whether to include the year field. Defaults to true.",
				},
			},
		},
		"date_time": schema.SingleNestedBlock{
			Description: "A date and time question.",
			Attributes: map[string]schema.Attribute{
				"question_text": schema.StringAttribute{
					Required:    true,
					Description: "The question text.",
				},
				"required": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(false),
					Description: "Whether the question is required.",
				},
				"include_year": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(true),
					Description: "Whether to include the year field. Defaults to true.",
				},
			},
		},
		"scale": schema.SingleNestedBlock{
			Description: "A linear scale question.",
			Attributes: map[string]schema.Attribute{
				"question_text": schema.StringAttribute{
					Required:    true,
					Description: "The question text.",
				},
				"required": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(false),
					Description: "Whether the question is required.",
				},
				"low": schema.Int64Attribute{
					Optional:    true,
					Description: "The lowest value on the scale. Defaults to 1.",
				},
				"high": schema.Int64Attribute{
					Optional:    true,
					Description: "The highest value on the scale. Defaults to 5.",
				},
				"low_label": schema.StringAttribute{
					Optional:    true,
					Description: "Label for the lowest value.",
				},
				"high_label": schema.StringAttribute{
					Optional:    true,
					Description: "Label for the highest value.",
				},
			},
		},
		"section_header": schema.SingleNestedBlock{
			Description: "A section header / page break. Has title and description but no question.",
			Attributes: map[string]schema.Attribute{
				"title": schema.StringAttribute{
					Required:    true,
					Description: "The section title.",
				},
				"description": schema.StringAttribute{
					Optional:    true,
					Description: "The section description.",
				},
			},
		},
	}
}

// gradingBlockSchema returns the grading SingleNestedBlock definition.
func gradingBlockSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Quiz grading options. Requires quiz = true on the form.",
		Attributes: map[string]schema.Attribute{
			"points": schema.Int64Attribute{
				Required:    true,
				Description: "Point value for this question.",
			},
			"correct_answer": schema.StringAttribute{
				Optional:    true,
				Description: "The correct answer. Must match an option value for multiple choice.",
			},
			"feedback_correct": schema.StringAttribute{
				Optional:    true,
				Description: "Feedback shown when the answer is correct.",
			},
			"feedback_incorrect": schema.StringAttribute{
				Optional:    true,
				Description: "Feedback shown when the answer is incorrect.",
			},
		},
	}
}
