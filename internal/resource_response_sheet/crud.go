// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceresponsesheet

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

// Create verifies that the form and spreadsheet exist, then stores the
// association in Terraform state. The Google Forms REST API v1 does not
// support programmatic linking of response destinations; the actual
// linking must be done manually in the Forms UI or via Apps Script.
func (r *ResponseSheetResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan ResponseSheetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	formID := plan.FormID.ValueString()
	spreadsheetID := plan.SpreadsheetID.ValueString()
	mode := "track"
	if !plan.Mode.IsNull() && !plan.Mode.IsUnknown() && plan.Mode.ValueString() != "" {
		mode = plan.Mode.ValueString()
	}

	// Step 1: Verify the form exists.
	tflog.Debug(ctx, "verifying form exists", map[string]interface{}{
		"form_id": formID,
	})

	f, err := r.client.Forms.Get(ctx, formID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Form Not Found",
			fmt.Sprintf("Could not find form %s: %s", formID, err.Error()),
		)
		return
	}

	// Step 2: Verify the spreadsheet exists and capture its URL.
	tflog.Debug(ctx, "verifying spreadsheet exists", map[string]interface{}{
		"spreadsheet_id": spreadsheetID,
	})

	ss, err := r.client.Sheets.Get(ctx, spreadsheetID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Spreadsheet Not Found",
			fmt.Sprintf("Could not find spreadsheet %s: %s", spreadsheetID, err.Error()),
		)
		return
	}

	linkedSheetID := strings.TrimSpace(f.LinkedSheetId)
	plan.LinkedSheetID = types.StringValue(linkedSheetID)
	plan.Linked = types.BoolValue(linkedSheetID != "" && linkedSheetID == spreadsheetID)
	plan.Mode = types.StringValue(mode)

	if mode == "validate" && !plan.Linked.ValueBool() {
		resp.Diagnostics.AddError(
			"Response Destination Not Linked",
			fmt.Sprintf("Form %s is not linked to spreadsheet %s (linkedSheetId=%q). Link responses in the Google Forms UI, then re-apply.", formID, spreadsheetID, linkedSheetID),
		)
		return
	}

	// Step 3: Build state.
	plan.ID = types.StringValue(fmt.Sprintf("%s#%s", formID, spreadsheetID))
	plan.SpreadsheetURL = types.StringValue(ss.SpreadsheetUrl)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	tflog.Info(ctx, "response sheet link created", map[string]interface{}{
		"form_id":        formID,
		"spreadsheet_id": spreadsheetID,
	})
	if mode == "track" {
		tflog.Warn(ctx, "response destination must be configured manually in Google Forms UI or via Apps Script")
	}
}

// Read verifies that the associated form and spreadsheet still exist.
// If either resource returns a 404, the link is removed from state.
func (r *ResponseSheetResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state ResponseSheetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mode := "track"
	if !state.Mode.IsNull() && !state.Mode.IsUnknown() && state.Mode.ValueString() != "" {
		mode = state.Mode.ValueString()
	}

	// Parse the composite ID.
	formID, spreadsheetID, err := parseCompositeID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Resource ID",
			fmt.Sprintf("Could not parse composite ID %q: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "reading response sheet link", map[string]interface{}{
		"form_id":        formID,
		"spreadsheet_id": spreadsheetID,
	})

	// Step 1: Verify the form still exists.
	f, err := r.client.Forms.Get(ctx, formID)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Google Form not found, removing response sheet link from state", map[string]interface{}{
				"form_id": formID,
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Google Form",
			fmt.Sprintf("Could not read form %s: %s", formID, err.Error()),
		)
		return
	}

	// Step 2: Verify the spreadsheet still exists and refresh the URL.
	ss, err := r.client.Sheets.Get(ctx, spreadsheetID)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Google Spreadsheet not found, removing response sheet link from state", map[string]interface{}{
				"spreadsheet_id": spreadsheetID,
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Google Spreadsheet",
			fmt.Sprintf("Could not read spreadsheet %s: %s", spreadsheetID, err.Error()),
		)
		return
	}

	// Step 3: Refresh computed fields.
	state.FormID = types.StringValue(formID)
	state.SpreadsheetID = types.StringValue(spreadsheetID)
	state.Mode = types.StringValue(mode)
	state.SpreadsheetURL = types.StringValue(ss.SpreadsheetUrl)

	linkedSheetID := strings.TrimSpace(f.LinkedSheetId)
	state.LinkedSheetID = types.StringValue(linkedSheetID)
	state.Linked = types.BoolValue(linkedSheetID != "" && linkedSheetID == spreadsheetID)

	if mode == "validate" && !state.Linked.ValueBool() {
		resp.Diagnostics.AddError(
			"Response Destination Not Linked",
			fmt.Sprintf("Form %s is not linked to spreadsheet %s (linkedSheetId=%q). Link responses in the Google Forms UI, then re-apply.", formID, spreadsheetID, linkedSheetID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is not expected to be called because both form_id and spreadsheet_id
// have RequiresReplace plan modifiers. Terraform will destroy and recreate
// the resource instead of updating it.
func (r *ResponseSheetResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// This should never be reached due to ForceNew on all configurable
	// attributes. If it is reached, return an error to surface the issue.
	resp.Diagnostics.AddError(
		"Unexpected Update",
		"googleforms_response_sheet does not support in-place updates. "+
			"Both form_id and spreadsheet_id require replacement.",
	)
}

// Delete is a no-op because the Google Forms REST API v1 does not support
// programmatic unlinking of response destinations. The resource is simply
// removed from Terraform state.
func (r *ResponseSheetResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state ResponseSheetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "removing response sheet link from state (no-op)", map[string]interface{}{
		"form_id":        state.FormID.ValueString(),
		"spreadsheet_id": state.SpreadsheetID.ValueString(),
	})

	// State is automatically removed by the framework after Delete returns
	// without errors.
}

// parseCompositeID splits a "formID#spreadsheetID" composite ID into its
// two constituent parts.
func parseCompositeID(id string) (formID string, spreadsheetID string, err error) {
	parts := strings.SplitN(id, "#", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("expected format formID#spreadsheetID, got %q", id)
	}

	return parts[0], parts[1], nil
}
