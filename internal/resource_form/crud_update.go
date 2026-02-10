// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"
	"fmt"

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

	// Step 1: Fetch current form to know the existing item count.
	currentForm, err := r.client.Forms.Get(ctx, formID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Google Form Before Update",
			fmt.Sprintf("Could not read form %s: %s", formID, err),
		)
		return
	}

	// Step 2: Build batch update requests.
	// Order: (1) UpdateFormInfo, (2) delete items, (3) quiz settings, (4) create items.
	// Item deletes MUST precede quiz settings changes so that disabling quiz
	// mode does not fail due to still-existing graded items.
	var requests []*forms.Request

	// Always update title and description.
	requests = append(requests, convert.BuildUpdateInfoRequest(
		plan.Title.ValueString(),
		plan.Description.ValueString(),
	))

	// Step 3: Handle item changes using replace-all strategy.
	// Delete all existing items, then create all items from plan.
	existingItemCount := len(currentForm.Items)

	// Collect create-item requests separately so we can insert quiz settings
	// between deletes and creates.
	var createItemRequests []*forms.Request

	if !plan.ContentJSON.IsNull() && !plan.ContentJSON.IsUnknown() && plan.ContentJSON.ValueString() != "" {
		// content_json mode
		tflog.Debug(ctx, "updating items via content_json mode")

		if existingItemCount > 0 {
			requests = append(requests, convert.BuildDeleteRequests(existingItemCount)...)
		}

		jsonRequests, err := convert.DeclarativeJSONToRequests(plan.ContentJSON.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing content_json",
				fmt.Sprintf("Could not parse content_json: %s", err),
			)
			return
		}
		createItemRequests = jsonRequests
	} else {
		// HCL item blocks mode
		convertItems, diags := tfItemsToConvertItems(ctx, plan.Items)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Only touch items if there are items in the plan or existing items to delete.
		if len(convertItems) > 0 || existingItemCount > 0 {
			tflog.Debug(ctx, "replacing items", map[string]interface{}{
				"existing_count": existingItemCount,
				"new_count":      len(convertItems),
			})

			// Delete existing items first (in reverse order).
			if existingItemCount > 0 {
				requests = append(requests, convert.BuildDeleteRequests(existingItemCount)...)
			}

			// Create new items from plan.
			if len(convertItems) > 0 {
				itemRequests, err := convert.ItemsToCreateRequests(convertItems)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error Building Item Requests",
						fmt.Sprintf("Could not build item create requests: %s", err),
					)
					return
				}
				createItemRequests = itemRequests
			}
		}
	}

	// Update quiz settings if changed (after deletes, before creates).
	planQuiz := plan.Quiz.ValueBool()
	stateQuiz := state.Quiz.ValueBool()
	if planQuiz != stateQuiz {
		tflog.Debug(ctx, "updating quiz settings", map[string]interface{}{
			"quiz": planQuiz,
		})
		requests = append(requests, convert.BuildQuizSettingsRequest(planQuiz))
	}

	// Append create-item requests after quiz settings.
	requests = append(requests, createItemRequests...)

	// Step 4: Execute batchUpdate.
	if len(requests) > 0 {
		tflog.Debug(ctx, "executing batchUpdate", map[string]interface{}{
			"request_count": len(requests),
		})

		batchReq := &forms.BatchUpdateFormRequest{
			Requests:              requests,
			IncludeFormInResponse: true,
		}

		_, err := r.client.Forms.BatchUpdate(ctx, formID, batchReq)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Google Form",
				fmt.Sprintf("BatchUpdate failed for form %s: %s", formID, err),
			)
			return
		}
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

	// Step 7: Build key map for item correlation.
	// For replace-all, items are new so correlate by position using plan items.
	// ASSUMPTION: The Google Forms API returns items in the same order
	// they were created via batchUpdate. If the API ever reorders items,
	// this positional mapping will produce incorrect item_key assignments.
	var keyMap map[string]string
	if !plan.Items.IsNull() && !plan.Items.IsUnknown() {
		var planItems []ItemModel
		diags := plan.Items.ElementsAs(ctx, &planItems, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		keyMap = make(map[string]string)
		for i, apiItem := range finalForm.Items {
			if i < len(planItems) {
				keyMap[apiItem.ItemId] = planItems[i].ItemKey.ValueString()
			}
		}
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
