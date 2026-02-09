// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/your-org/terraform-provider-googleforms/internal/client"
	"github.com/your-org/terraform-provider-googleforms/internal/convert"
)

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

	// Step 5: Map to Terraform state, preserving plan/config values.
	newState := convertFormModelToTFState(formModel, state)

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
