// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	forms "google.golang.org/api/forms/v1"

	"github.com/45ck/terraform-provider-googleforms/internal/convert"
)

// Update replaces the form's settings and items with the planned configuration.
func (r *FormResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan FormResourceModel
	var state FormResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	formID := state.ID.ValueString()
	tflog.Debug(ctx, "updating Google Form", map[string]interface{}{
		"form_id": formID,
	})

	updateStrategy := "replace_all"
	if !plan.UpdateStrategy.IsNull() && !plan.UpdateStrategy.IsUnknown() && plan.UpdateStrategy.ValueString() != "" {
		updateStrategy = plan.UpdateStrategy.ValueString()
	}
	dangerReplaceAll := false
	if !plan.DangerousReplaceAll.IsNull() && !plan.DangerousReplaceAll.IsUnknown() {
		dangerReplaceAll = plan.DangerousReplaceAll.ValueBool()
	}

	// Step 1: Fetch current form to know the existing item count.
	currentForm, err := r.client.Forms.Get(ctx, formID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Google Form Before Update",
			fmt.Sprintf("Could not read form %s: %s", formID, err),
		)
		return
	}

	// Step 2: Build and execute batch update requests.
	// Always update title/description. Items are updated based on update_strategy.
	switch updateStrategy {
	case "targeted":
		resp.Diagnostics.Append(r.updateTargeted(ctx, plan, state, currentForm)...)
		if resp.Diagnostics.HasError() {
			return
		}
	case "replace_all":
		if !dangerReplaceAll {
			resp.Diagnostics.AddWarning(
				"Replace-All Item Updates Enabled",
				"update_strategy is 'replace_all'. This deletes and recreates all items on changes, which can break response mappings and integrations. Set update_strategy = \"targeted\" where possible, or set dangerously_replace_all_items = true to acknowledge this behavior.",
			)
		}

		resp.Diagnostics.Append(r.updateReplaceAll(ctx, plan, state, currentForm)...)
		if resp.Diagnostics.HasError() {
			return
		}
	default:
		resp.Diagnostics.AddError("Invalid update_strategy", fmt.Sprintf("Unsupported update_strategy %q", updateStrategy))
		return
	}

	// Step 5: Handle publish settings changes.
	planPublished := plan.Published.ValueBool()
	planAccepting := plan.AcceptingResponses.ValueBool()
	statePublished := state.Published.ValueBool()
	stateAccepting := state.AcceptingResponses.ValueBool()

	if planPublished != statePublished || planAccepting != stateAccepting {
		tflog.Debug(ctx, "updating publish settings", map[string]interface{}{
			"published":           planPublished,
			"accepting_responses": planAccepting,
		})

		err := r.client.Forms.SetPublishSettings(ctx, formID, planPublished, planAccepting)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Publish Settings",
				fmt.Sprintf("Could not update publish settings for form %s: %s", formID, err),
			)
			return
		}
	}

	// Step 6: Read back the final form state.
	finalForm, err := r.client.Forms.Get(ctx, formID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Google Form After Update",
			fmt.Sprintf("Could not read form %s after update: %s", formID, err),
		)
		return
	}

	// Step 7: Build key map for item correlation from prior state to keep item_key stable.
	keyMap, diags := buildItemKeyMap(ctx, state.Items)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	formModel, err := convert.FormToModel(finalForm, keyMap)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Converting Form Response",
			fmt.Sprintf("Could not convert API response for form %s: %s", formID, err),
		)
		return
	}

	// Step 8: Convert to TF state and save.
	newState := convertFormModelToTFState(formModel, plan)

	if plan.ContentJSON.IsNull() || plan.ContentJSON.IsUnknown() || plan.ContentJSON.ValueString() == "" {
		itemList, diags := convertItemsToTFList(ctx, formModel.Items)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		newState.Items = itemList
	} else {
		newState.Items = plan.Items
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *FormResource) updateReplaceAll(
	ctx context.Context,
	plan FormResourceModel,
	state FormResourceModel,
	currentForm *forms.Form,
) diag.Diagnostics {
	var diags diag.Diagnostics

	// Order: (1) UpdateFormInfo, (2) delete items, (3) quiz settings, (4) create items.
	// Item deletes MUST precede quiz settings changes so that disabling quiz
	// mode does not fail due to still-existing graded items.
	var requests []*forms.Request

	requests = append(requests, convert.BuildUpdateInfoRequest(
		plan.Title.ValueString(),
		plan.Description.ValueString(),
	))

	existingItemCount := len(currentForm.Items)
	var createItemRequests []*forms.Request

	if !plan.ContentJSON.IsNull() && !plan.ContentJSON.IsUnknown() && plan.ContentJSON.ValueString() != "" {
		tflog.Debug(ctx, "updating items via content_json mode")

		if existingItemCount > 0 {
			requests = append(requests, convert.BuildDeleteRequests(existingItemCount)...)
		}

		jsonRequests, err := convert.DeclarativeJSONToRequests(plan.ContentJSON.ValueString())
		if err != nil {
			diags.AddError("Error Parsing content_json", fmt.Sprintf("Could not parse content_json: %s", err))
			return diags
		}
		createItemRequests = jsonRequests
	} else {
		convertItems, d := tfItemsToConvertItems(ctx, plan.Items)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		if len(convertItems) > 0 || existingItemCount > 0 {
			tflog.Debug(ctx, "replacing items", map[string]interface{}{
				"existing_count": existingItemCount,
				"new_count":      len(convertItems),
			})

			if existingItemCount > 0 {
				requests = append(requests, convert.BuildDeleteRequests(existingItemCount)...)
			}

			if len(convertItems) > 0 {
				itemRequests, err := convert.ItemsToCreateRequests(convertItems)
				if err != nil {
					diags.AddError("Error Building Item Requests", fmt.Sprintf("Could not build item create requests: %s", err))
					return diags
				}
				createItemRequests = itemRequests
			}
		}
	}

	planQuiz := plan.Quiz.ValueBool()
	stateQuiz := state.Quiz.ValueBool()
	if planQuiz != stateQuiz {
		tflog.Debug(ctx, "updating quiz settings", map[string]interface{}{
			"quiz": planQuiz,
		})
		requests = append(requests, convert.BuildQuizSettingsRequest(planQuiz))
	}

	requests = append(requests, createItemRequests...)

	if len(requests) == 0 {
		return diags
	}

	tflog.Debug(ctx, "executing batchUpdate (replace_all)", map[string]interface{}{
		"request_count": len(requests),
	})

	batchReq := &forms.BatchUpdateFormRequest{
		Requests:              requests,
		IncludeFormInResponse: true,
	}

	_, err := r.client.Forms.BatchUpdate(ctx, state.ID.ValueString(), batchReq)
	if err != nil {
		diags.AddError("Error Updating Google Form", fmt.Sprintf("BatchUpdate failed for form %s: %s", state.ID.ValueString(), err))
		return diags
	}

	return diags
}

func (r *FormResource) updateTargeted(
	ctx context.Context,
	plan FormResourceModel,
	state FormResourceModel,
	currentForm *forms.Form,
) diag.Diagnostics {
	var diags diag.Diagnostics

	if !plan.ContentJSON.IsNull() && !plan.ContentJSON.IsUnknown() && plan.ContentJSON.ValueString() != "" {
		diags.AddError("Targeted Updates Not Supported With content_json", "content_json mode cannot be updated in-place. Set update_strategy = \"replace_all\" or switch to item blocks.")
		return diags
	}

	// Decode plan/state items to enforce "no structural changes" for targeted mode.
	var planItems []ItemModel
	var stateItems []ItemModel

	if !plan.Items.IsNull() && !plan.Items.IsUnknown() {
		diags.Append(plan.Items.ElementsAs(ctx, &planItems, false)...)
	}
	if !state.Items.IsNull() && !state.Items.IsUnknown() {
		diags.Append(state.Items.ElementsAs(ctx, &stateItems, false)...)
	}
	if diags.HasError() {
		return diags
	}

	if len(planItems) != len(stateItems) {
		diags.AddError(
			"Targeted Update Requires Stable Item Structure",
			"Targeted updates require the same number of items. Add/remove/reorder items using update_strategy = \"replace_all\".",
		)
		return diags
	}

	// Ensure the item_key ordering is unchanged, and that we have google_item_id for every item.
	for i := range planItems {
		if planItems[i].ItemKey.ValueString() != stateItems[i].ItemKey.ValueString() {
			diags.AddError(
				"Targeted Update Requires Stable Item Order",
				"Targeted updates require the same item_key ordering. Reordering items requires update_strategy = \"replace_all\".",
			)
			return diags
		}

		if stateItems[i].GoogleItemID.IsNull() || stateItems[i].GoogleItemID.IsUnknown() || stateItems[i].GoogleItemID.ValueString() == "" {
			diags.AddError(
				"Missing google_item_id",
				"Targeted updates require google_item_id to be known for all items. Re-import the form or switch to update_strategy = \"replace_all\".",
			)
			return diags
		}
	}

	desiredItems, d := tfItemsToConvertItems(ctx, plan.Items)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	// Index current items by itemId to find existing items quickly.
	type itemRef struct {
		index int64
		item  *forms.Item
	}
	byID := make(map[string]itemRef, len(currentForm.Items))
	for i, it := range currentForm.Items {
		if it != nil && it.ItemId != "" {
			byID[it.ItemId] = itemRef{index: int64(i), item: it}
		}
	}

	var requests []*forms.Request
	requests = append(requests, convert.BuildUpdateInfoRequest(
		plan.Title.ValueString(),
		plan.Description.ValueString(),
	))

	planQuiz := plan.Quiz.ValueBool()
	stateQuiz := state.Quiz.ValueBool()
	if planQuiz != stateQuiz {
		requests = append(requests, convert.BuildQuizSettingsRequest(planQuiz))
	}

	for i := range desiredItems {
		itemID := stateItems[i].GoogleItemID.ValueString()
		ref, ok := byID[itemID]
		if !ok || ref.item == nil {
			diags.AddError(
				"Targeted Update Failed To Locate Existing Item",
				fmt.Sprintf("Could not find existing item %q (google_item_id=%s) in the current form. Switch to update_strategy = \"replace_all\".", stateItems[i].ItemKey.ValueString(), itemID),
			)
			return diags
		}

		// Apply desired changes to the existing API object to preserve IDs.
		needsReplace, err := convert.ApplyItemModelToExistingItem(ref.item, desiredItems[i])
		if err != nil {
			diags.AddError("Targeted Update Failed", err.Error())
			return diags
		}
		if needsReplace {
			diags.AddError(
				"Targeted Update Requires Replace-All",
				fmt.Sprintf("Item %q requires a structural change (e.g. question type change) and cannot be updated in-place. Switch to update_strategy = \"replace_all\".", stateItems[i].ItemKey.ValueString()),
			)
			return diags
		}

		requests = append(requests, &forms.Request{
			UpdateItem: &forms.UpdateItemRequest{
				Item:       ref.item,
				Location:   &forms.Location{Index: ref.index},
				UpdateMask: "*",
			},
		})
	}

	if len(requests) == 0 {
		return diags
	}

	tflog.Debug(ctx, "executing batchUpdate (targeted)", map[string]interface{}{
		"request_count": len(requests),
	})

	batchReq := &forms.BatchUpdateFormRequest{
		Requests:              requests,
		IncludeFormInResponse: true,
	}

	_, err := r.client.Forms.BatchUpdate(ctx, state.ID.ValueString(), batchReq)
	if err != nil {
		diags.AddError("Error Updating Google Form", fmt.Sprintf("BatchUpdate failed for form %s: %s", state.ID.ValueString(), err))
		return diags
	}

	return diags
}
