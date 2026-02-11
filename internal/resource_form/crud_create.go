// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	forms "google.golang.org/api/forms/v1"

	"github.com/45ck/terraform-provider-googleforms/internal/convert"
)

// Create creates a new Google Form with the configured items and settings.
func (r *FormResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan FormResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Step 1: Create the form with title only (API limitation).
	// The Google Forms API only accepts Info.Title during creation;
	// all other fields must be set via batchUpdate.
	tflog.Debug(ctx, "creating Google Form", map[string]interface{}{
		"title": plan.Title.ValueString(),
	})

	createForm := &forms.Form{
		Info: &forms.Info{
			Title: plan.Title.ValueString(),
		},
	}

	result, err := r.client.Forms.Create(ctx, createForm)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Google Form",
			fmt.Sprintf("Could not create form: %s", err),
		)
		return
	}

	tflog.Info(ctx, "created Google Form", map[string]interface{}{
		"form_id": result.FormId,
	})

	// Step 2: CRITICAL partial state save. Persist the form ID immediately
	// so the resource can be tracked even if subsequent API calls fail.
	// This prevents orphaned forms that exist in Google but not in state.
	plan.ID = types.StringValue(result.FormId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	formID := result.FormId

	// Step 3: Build batch update requests for all settings and items.
	var requests []*forms.Request
	var createKeys []string
	var desiredItems []convert.ItemModel

	// Always send title+description via batchUpdate to ensure description is set.
	description := plan.Description.ValueString()
	requests = append(requests, convert.BuildUpdateInfoRequest(plan.Title.ValueString(), description))

	// Enable quiz mode if requested.
	if plan.Quiz.ValueBool() {
		requests = append(requests, convert.BuildQuizSettingsRequest(true))
	}

	// Optional: email collection type.
	if !plan.EmailCollectionType.IsNull() && !plan.EmailCollectionType.IsUnknown() && plan.EmailCollectionType.ValueString() != "" {
		requests = append(requests, convert.BuildEmailCollectionTypeRequest(plan.EmailCollectionType.ValueString()))
	}

	// Add items from either content_json or item blocks.
	if !plan.ContentJSON.IsNull() && !plan.ContentJSON.IsUnknown() && plan.ContentJSON.ValueString() != "" {
		tflog.Debug(ctx, "using content_json mode for items")
		jsonRequests, err := convert.DeclarativeJSONToRequests(plan.ContentJSON.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing content_json",
				fmt.Sprintf("Could not parse content_json: %s", err),
			)
			return
		}
		requests = append(requests, jsonRequests...)
	} else {
		convertItems, diags := tfItemsToConvertItems(ctx, plan.Items)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		desiredItems = convertItems

		if len(convertItems) > 0 {
			var planItems []ItemModel
			diags := plan.Items.ElementsAs(ctx, &planItems, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			createKeys = make([]string, len(planItems))
			for i := range planItems {
				createKeys[i] = planItems[i].ItemKey.ValueString()
			}

			tflog.Debug(ctx, "creating items from HCL blocks", map[string]interface{}{
				"item_count": len(convertItems),
			})
			itemRequests, err := convert.ItemsToCreateRequests(convertItems)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Building Item Requests",
					fmt.Sprintf("Could not build item create requests: %s", err),
				)
				return
			}
			requests = append(requests, itemRequests...)
		}
	}

	// Step 4: Execute batchUpdate if there are any requests.
	var batchResp *forms.BatchUpdateFormResponse
	if len(requests) > 0 {
		tflog.Debug(ctx, "executing batchUpdate", map[string]interface{}{
			"request_count": len(requests),
		})

		batchReq := &forms.BatchUpdateFormRequest{
			Requests:              requests,
			IncludeFormInResponse: true,
		}

		apiResp, err := r.client.Forms.BatchUpdate(ctx, formID, batchReq)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Google Form",
				fmt.Sprintf("Form was created (ID: %s) but batchUpdate failed: %s", formID, err),
			)
			return
		}
		batchResp = apiResp
	}

	// Step 5: Set publish settings if published or accepting_responses is true.
	published := plan.Published.ValueBool()
	accepting := plan.AcceptingResponses.ValueBool()
	if published || accepting {
		tflog.Debug(ctx, "setting publish settings", map[string]interface{}{
			"published":           published,
			"accepting_responses": accepting,
		})

		err := r.client.Forms.SetPublishSettings(ctx, formID, published, accepting)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Publish Settings",
				fmt.Sprintf("Form was created (ID: %s) but publish settings failed: %s", formID, err),
			)
			return
		}
	}

	// Step 5b: Optional: move the form into a Drive folder.
	supportsAllDrives := false
	if !plan.SupportsAllDrives.IsNull() && !plan.SupportsAllDrives.IsUnknown() {
		supportsAllDrives = plan.SupportsAllDrives.ValueBool()
	}
	if !plan.FolderID.IsNull() && !plan.FolderID.IsUnknown() && plan.FolderID.ValueString() != "" {
		if err := r.client.Drive.MoveToFolder(ctx, formID, plan.FolderID.ValueString(), supportsAllDrives); err != nil {
			resp.Diagnostics.AddError("Move Form To Folder Failed", err.Error())
			return
		}
	}
	// Best-effort: record current parents.
	if parents, err := r.client.Drive.GetParents(ctx, formID, supportsAllDrives); err == nil {
		lv, diags := types.ListValueFrom(ctx, types.StringType, parents)
		resp.Diagnostics.Append(diags...)
		plan.ParentIDs = lv
	} else {
		resp.Diagnostics.AddWarning("Drive Parents Unavailable", err.Error())
		plan.ParentIDs = types.ListNull(types.StringType)
	}

	// Step 6: Read back the final form state to capture all computed fields.
	finalForm, err := r.client.Forms.Get(ctx, formID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Google Form After Create",
			fmt.Sprintf("Form was created (ID: %s) but could not read back final state: %s", formID, err),
		)
		return
	}

	// Step 7: Convert API response to Terraform state.
	// Build key map from plan items so item_keys are preserved.
	keyMap, diags := buildItemKeyMap(ctx, plan.Items)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(createKeys) > 0 && batchResp != nil {
		createReqIndexToKey, derr := buildCreateReqIndexToKey(requests, createKeys)
		if derr != nil {
			resp.Diagnostics.AddWarning("Item Key Correlation Failed", derr.Error())
		} else {
			createdKeyMap, derr := extractCreateItemKeyMap(batchResp, requests, createReqIndexToKey)
			if derr != nil {
				resp.Diagnostics.AddWarning("Item Key Correlation Failed", derr.Error())
			} else {
				keyMap = createdKeyMap
			}
		}
	}

	// Fallback: correlate by position if reply correlation was unavailable.
	if keyMap == nil && !plan.Items.IsNull() && !plan.Items.IsUnknown() {
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

	// Step 7b: Apply choice navigation updates (go_to_section_key/go_to_section_id).
	if len(desiredItems) > 0 && keyMap != nil {
		diags := r.applyChoiceNavigationUpdates(ctx, formID, desiredItems, keyMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Re-read after applying navigation updates so state reflects them.
		finalForm, err = r.client.Forms.Get(ctx, formID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Google Form After Navigation Update",
				fmt.Sprintf("Form was created (ID: %s) but could not read back final state: %s", formID, err),
			)
			return
		}
	}

	formModel, err := convert.FormToModel(finalForm, keyMap)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Converting Form Response",
			fmt.Sprintf("Could not convert API response to state: %s", err),
		)
		return
	}

	// Preserve input-only fields that may not be returned by the API.
	formModel.Items, diags = overlayConvertItemInputsFromTF(ctx, formModel.Items, plan.Items)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := convertFormModelToTFState(formModel, plan)

	// Set items in state (unless using content_json mode).
	if plan.ContentJSON.IsNull() || plan.ContentJSON.IsUnknown() || plan.ContentJSON.ValueString() == "" {
		itemList, diags := convertItemsToTFList(ctx, formModel.Items)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Items = itemList
	} else {
		// In content_json mode, keep the items list null.
		state.Items = plan.Items
	}

	// Step 8: Save final state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
