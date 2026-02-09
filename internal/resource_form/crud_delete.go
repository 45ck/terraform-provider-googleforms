// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/your-org/terraform-provider-googleforms/internal/client"
)

func (r *FormResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state FormResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	formID := state.ID.ValueString()
	tflog.Debug(ctx, "deleting Google Form", map[string]interface{}{
		"form_id": formID,
	})

	// Delete the form via the Drive API. The Drive.Delete interface
	// already treats 404 as success (returns nil), but we guard
	// against implementations that may not.
	err := r.client.Drive.Delete(ctx, formID)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Google Form already deleted", map[string]interface{}{
				"form_id": formID,
			})
			return
		}

		resp.Diagnostics.AddError(
			"Error Deleting Google Form",
			fmt.Sprintf("Could not delete form %s: %s", formID, err),
		)
		return
	}

	tflog.Info(ctx, "deleted Google Form", map[string]interface{}{
		"form_id": formID,
	})

	// State is automatically removed by the framework after Delete returns
	// without errors.
}
