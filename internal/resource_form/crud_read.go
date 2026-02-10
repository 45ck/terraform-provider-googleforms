// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
	"github.com/45ck/terraform-provider-googleforms/internal/convert"
)

// Read fetches the current state of a Google Form from the API.
func (r *FormResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state FormResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	formID := state.ID.ValueString()
	tflog.Debug(ctx, "reading Google Form", map[string]interface{}{
		"form_id": formID,
	})

	// Step 1: Fetch the form from the API.
	form, err := r.client.Forms.Get(ctx, formID)
	if err != nil {
		// Step 2: Handle 404 - form deleted outside Terraform.
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Google Form not found, removing from state", map[string]interface{}{
				"form_id": formID,
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Google Form",
			fmt.Sprintf("Could not read form %s: %s", formID, err),
		)
		return
	}

	// Step 3: Build key map from current state items for item_key correlation.
	keyMap, diags := buildItemKeyMap(ctx, state.Items)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Step 4: Convert API response to convert.FormModel.
	formModel, err := convert.FormToModel(form, keyMap)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Converting Form Response",
			fmt.Sprintf("Could not convert API response for form %s: %s", formID, err),
		)
		return
	}

	// In partial management mode, only keep items that are already tracked in
	// state (keyMap). Unmanaged items remain in the form but are ignored by TF.
	if state.ManageMode.ValueString() == "partial" {
		formModel.Items = filterItemsByKeyMap(formModel.Items, keyMap)
	}

	// Preserve input-only fields that may not be returned by the API.
	formModel.Items, diags = overlayConvertItemInputsFromTF(ctx, formModel.Items, state.Items)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Step 5: Map to Terraform state, preserving plan/config values.
	newState := convertFormModelToTFState(formModel, state)

	// Best-effort: record current Drive parents.
	supportsAllDrives := false
	if !state.SupportsAllDrives.IsNull() && !state.SupportsAllDrives.IsUnknown() {
		supportsAllDrives = state.SupportsAllDrives.ValueBool()
	}
	if parents, err := r.client.Drive.GetParents(ctx, formID, supportsAllDrives); err == nil {
		lv, diags := types.ListValueFrom(ctx, types.StringType, parents)
		resp.Diagnostics.Append(diags...)
		newState.ParentIDs = lv
	} else {
		resp.Diagnostics.AddWarning("Drive Parents Unavailable", err.Error())
		newState.ParentIDs = types.ListNull(types.StringType)
	}

	// Step 6: Set items (unless using content_json mode).
	if state.ContentJSON.IsNull() || state.ContentJSON.IsUnknown() || state.ContentJSON.ValueString() == "" {
		itemList, diags := convertItemsToTFList(ctx, formModel.Items)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		newState.Items = itemList
	} else {
		// In content_json mode, preserve the existing content_json value.
		// Drift detection is hash-based via the plan modifier.
		newState.ContentJSON = state.ContentJSON
		newState.Items = state.Items
	}

	// Step 7: Save to state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}
