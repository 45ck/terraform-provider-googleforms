// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Compile-time interface checks.
var (
	_ resource.ConfigValidator = MutuallyExclusiveValidator{}
	_ resource.ConfigValidator = AcceptingResponsesRequiresPublishedValidator{}
	_ resource.ConfigValidator = UniqueItemKeyValidator{}
	_ resource.ConfigValidator = ExactlyOneSubBlockValidator{}
	_ resource.ConfigValidator = OptionsRequiredForChoiceValidator{}
	_ resource.ConfigValidator = CorrectAnswerInOptionsValidator{}
	_ resource.ConfigValidator = GradingRequiresQuizValidator{}
)

// ---------------------------------------------------------------------------
// 1. MutuallyExclusiveValidator
// ---------------------------------------------------------------------------

// MutuallyExclusiveValidator ensures content_json and item blocks are not
// both specified on the same resource.
type MutuallyExclusiveValidator struct{}

func (v MutuallyExclusiveValidator) Description(_ context.Context) string {
	return "Validates that content_json and item blocks are not both set."
}

func (v MutuallyExclusiveValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v MutuallyExclusiveValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var contentJSON types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("content_json"), &contentJSON)...)

	var items types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("item"), &items)...)

	if resp.Diagnostics.HasError() {
		return
	}

	hasContentJSON := !contentJSON.IsNull() && !contentJSON.IsUnknown()
	hasItems := !items.IsNull() && !items.IsUnknown() && len(items.Elements()) > 0

	if hasContentJSON && hasItems {
		resp.Diagnostics.AddError(
			"Conflicting Configuration",
			`Cannot use both "content_json" and "item" blocks in the same form resource.`,
		)
	}
}

// ---------------------------------------------------------------------------
// 2. AcceptingResponsesRequiresPublishedValidator
// ---------------------------------------------------------------------------

// AcceptingResponsesRequiresPublishedValidator ensures accepting_responses is
// only true when published is also true.
type AcceptingResponsesRequiresPublishedValidator struct{}

func (v AcceptingResponsesRequiresPublishedValidator) Description(_ context.Context) string {
	return "Validates that accepting_responses=true requires published=true."
}

func (v AcceptingResponsesRequiresPublishedValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v AcceptingResponsesRequiresPublishedValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var accepting types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("accepting_responses"), &accepting)...)

	var published types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("published"), &published)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if accepting.IsNull() || accepting.IsUnknown() || !accepting.ValueBool() {
		return
	}

	if published.IsNull() || published.IsUnknown() || !published.ValueBool() {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"A form cannot accept responses while unpublished. "+
				"Set published = true or remove accepting_responses.",
		)
	}
}

// ---------------------------------------------------------------------------
// 3. UniqueItemKeyValidator
// ---------------------------------------------------------------------------

// UniqueItemKeyValidator ensures all item_key values are unique.
type UniqueItemKeyValidator struct{}

func (v UniqueItemKeyValidator) Description(_ context.Context) string {
	return "Validates that all item_key values within the form are unique."
}

func (v UniqueItemKeyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v UniqueItemKeyValidator) ValidateResource(
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

	seen := make(map[string]bool, len(itemModels))
	for _, item := range itemModels {
		key := item.ItemKey.ValueString()
		if key == "" || item.ItemKey.IsNull() || item.ItemKey.IsUnknown() {
			continue
		}
		if seen[key] {
			resp.Diagnostics.AddError(
				"Duplicate Item Key",
				`Duplicate item_key "`+key+`" found. `+
					"Each item_key must be unique within a form resource.",
			)
			return
		}
		seen[key] = true
	}
}

// ---------------------------------------------------------------------------
// 4. ExactlyOneSubBlockValidator
// ---------------------------------------------------------------------------

// ExactlyOneSubBlockValidator ensures each item has exactly one question type
// sub-block (multiple_choice, short_answer, or paragraph).
type ExactlyOneSubBlockValidator struct{}

func (v ExactlyOneSubBlockValidator) Description(_ context.Context) string {
	return "Validates that each item has exactly one question type sub-block."
}

func (v ExactlyOneSubBlockValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ExactlyOneSubBlockValidator) ValidateResource(
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

	for i, item := range itemModels {
		count := countSubBlocks(item)
		if count != 1 {
			identity := fmt.Sprintf("index %d", i)
			if key := item.ItemKey.ValueString(); key != "" {
				identity = fmt.Sprintf("%q", key)
			}
			resp.Diagnostics.AddError(
				"Invalid Item Configuration",
				fmt.Sprintf(
					"Item %s must have exactly one question type "+
						"(multiple_choice, short_answer, or paragraph), but has %d.",
					identity, count,
				),
			)
			return
		}
	}
}

// countSubBlocks returns how many question type sub-blocks are non-nil.
func countSubBlocks(item ItemModel) int {
	count := 0
	if item.MultipleChoice != nil {
		count++
	}
	if item.ShortAnswer != nil {
		count++
	}
	if item.Paragraph != nil {
		count++
	}
	return count
}
