// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
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
		Description: "Manages a Google Form. Note: some Forms item types are not supported by the API for creation (for example file upload questions). Use content_json as an escape hatch when needed.",
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
		"manage_mode": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("all"),
			Description: "Management mode for items. 'all' treats the item list as authoritative for the whole form. 'partial' only manages the configured items (by item_key) and leaves other items untouched; in partial mode, new items are appended by default.",
			Validators: []validator.String{
				stringvalidator.OneOf("all", "partial"),
			},
		},
		"partial_new_item_policy": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString("append"),
			Description: "Policy for placing newly created items when manage_mode = \"partial\". " +
				"'append' (default) adds new managed items to the end of the form without shifting unmanaged items. " +
				"'plan_index' inserts at the index specified by the plan's item list, which may shift unmanaged items.",
			Validators: []validator.String{
				stringvalidator.OneOf("append", "plan_index"),
			},
		},
		"conflict_policy": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("overwrite"),
			Description: "Conflict policy when the form was edited out-of-band. 'overwrite' applies changes to the latest revision. 'fail' uses write control (requiredRevisionId) and errors if the revision_id has changed since last read.",
			Validators: []validator.String{
				stringvalidator.OneOf("overwrite", "fail"),
			},
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
		"revision_id": schema.StringAttribute{
			Computed:    true,
			Description: "The form revision ID returned by the API (valid for ~24h). Used for conflict detection when conflict_policy = \"fail\".",
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
					Computed:    true,
					Default:     int64default.StaticInt64(1),
					Description: "The lowest value on the scale. Defaults to 1.",
					Validators: []validator.Int64{
						int64validator.Between(1, 10),
					},
				},
				"high": schema.Int64Attribute{
					Optional:    true,
					Computed:    true,
					Default:     int64default.StaticInt64(5),
					Description: "The highest value on the scale. Defaults to 5.",
					Validators: []validator.Int64{
						int64validator.Between(2, 10),
					},
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
		"time": schema.SingleNestedBlock{
			Description: "A time or duration question.",
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
				"duration": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(false),
					Description: "If true, the question is an elapsed time duration. Otherwise it is a time of day.",
				},
			},
		},
		"rating": schema.SingleNestedBlock{
			Description: "A rating question (stars/hearts/thumbs).",
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
				"icon_type": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString("STAR"),
					Description: "The icon type (STAR, HEART, THUMB_UP).",
					Validators: []validator.String{
						stringvalidator.OneOf("STAR", "HEART", "THUMB_UP"),
					},
				},
				"rating_scale_level": schema.Int64Attribute{
					Optional:    true,
					Computed:    true,
					Default:     int64default.StaticInt64(5),
					Description: "The number of icons (e.g. 5).",
					Validators: []validator.Int64{
						int64validator.Between(1, 10),
					},
				},
			},
		},
		"text_item": schema.SingleNestedBlock{
			Description: "A text-only item (no question).",
			Attributes: map[string]schema.Attribute{
				"title": schema.StringAttribute{
					Required:    true,
					Description: "Title shown to respondents.",
				},
				"description": schema.StringAttribute{
					Optional:    true,
					Description: "Optional description shown to respondents.",
				},
			},
		},
		"image": schema.SingleNestedBlock{
			Description: "An image item (non-question). Note: source_uri may not be returned by the API; the provider preserves the configured value in state for drift-free plans.",
			Attributes: map[string]schema.Attribute{
				"title": schema.StringAttribute{
					Optional:    true,
					Description: "Optional item title shown above the image.",
				},
				"description": schema.StringAttribute{
					Optional:    true,
					Description: "Optional item description shown above the image.",
				},
				"source_uri": schema.StringAttribute{
					Required:    true,
					Description: "Input-only. The source URI used to insert the image.",
				},
				"alt_text": schema.StringAttribute{
					Optional:    true,
					Description: "Optional alt text read by screen readers.",
				},
				"content_uri": schema.StringAttribute{
					Computed:    true,
					Description: "Output-only. Temporary download URI returned by the API.",
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
			},
		},
		"video": schema.SingleNestedBlock{
			Description: "A video item (non-question).",
			Attributes: map[string]schema.Attribute{
				"title": schema.StringAttribute{
					Optional:    true,
					Description: "Optional item title shown above the video.",
				},
				"description": schema.StringAttribute{
					Optional:    true,
					Description: "Optional item description shown above the video.",
				},
				"youtube_uri": schema.StringAttribute{
					Required:    true,
					Description: "Required. A YouTube URI.",
				},
				"caption": schema.StringAttribute{
					Optional:    true,
					Description: "Optional caption displayed below the video.",
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
