// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcespreadsheet

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/45ck/terraform-provider-googleforms/internal/client"

	sheets "google.golang.org/api/sheets/v4"
)

// Create creates a new Google Sheets spreadsheet.
func (r *SpreadsheetResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SpreadsheetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spreadsheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: plan.Title.ValueString(),
		},
	}

	if !plan.Locale.IsNull() && !plan.Locale.IsUnknown() {
		spreadsheet.Properties.Locale = plan.Locale.ValueString()
	}
	if !plan.TimeZone.IsNull() && !plan.TimeZone.IsUnknown() {
		spreadsheet.Properties.TimeZone = plan.TimeZone.ValueString()
	}

	created, err := r.client.Sheets.Create(ctx, spreadsheet)
	if err != nil {
		resp.Diagnostics.AddError("Create Spreadsheet Failed", err.Error())
		return
	}

	// Partial state save: write ID immediately so Terraform can track the
	// resource even if a subsequent step fails.
	plan.ID = types.StringValue(created.SpreadsheetId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.URL = types.StringValue(created.SpreadsheetUrl)
	if created.Properties != nil {
		plan.Title = types.StringValue(created.Properties.Title)
		if created.Properties.Locale != "" {
			plan.Locale = types.StringValue(created.Properties.Locale)
		}
		if created.Properties.TimeZone != "" {
			plan.TimeZone = types.StringValue(created.Properties.TimeZone)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	tflog.Debug(ctx, "created spreadsheet", map[string]interface{}{"id": created.SpreadsheetId})
}

// Read refreshes the Terraform state from the Google Sheets API.
func (r *SpreadsheetResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state SpreadsheetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spreadsheet, err := r.client.Sheets.Get(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Spreadsheet Failed", err.Error())
		return
	}

	state.ID = types.StringValue(spreadsheet.SpreadsheetId)
	state.URL = types.StringValue(spreadsheet.SpreadsheetUrl)
	if spreadsheet.Properties != nil {
		state.Title = types.StringValue(spreadsheet.Properties.Title)
		if spreadsheet.Properties.Locale != "" {
			state.Locale = types.StringValue(spreadsheet.Properties.Locale)
		}
		if spreadsheet.Properties.TimeZone != "" {
			state.TimeZone = types.StringValue(spreadsheet.Properties.TimeZone)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies an existing Google Sheets spreadsheet via batchUpdate.
func (r *SpreadsheetResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan SpreadsheetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state SpreadsheetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				UpdateSpreadsheetProperties: &sheets.UpdateSpreadsheetPropertiesRequest{
					Properties: &sheets.SpreadsheetProperties{
						Title: plan.Title.ValueString(),
					},
					Fields: "title",
				},
			},
		},
	}

	// Add locale update if set.
	if !plan.Locale.IsNull() && !plan.Locale.IsUnknown() {
		batchReq.Requests[0].UpdateSpreadsheetProperties.Properties.Locale = plan.Locale.ValueString()
		batchReq.Requests[0].UpdateSpreadsheetProperties.Fields += ",locale"
	}
	if !plan.TimeZone.IsNull() && !plan.TimeZone.IsUnknown() {
		batchReq.Requests[0].UpdateSpreadsheetProperties.Properties.TimeZone = plan.TimeZone.ValueString()
		batchReq.Requests[0].UpdateSpreadsheetProperties.Fields += ",timeZone"
	}

	_, err := r.client.Sheets.BatchUpdate(ctx, state.ID.ValueString(), batchReq)
	if err != nil {
		resp.Diagnostics.AddError("Update Spreadsheet Failed", err.Error())
		return
	}

	plan.ID = state.ID
	plan.URL = state.URL

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes the Google Sheets spreadsheet via the Drive API.
func (r *SpreadsheetResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state SpreadsheetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use Drive API to delete the spreadsheet file.
	err := r.client.Drive.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Delete Spreadsheet Failed",
			fmt.Sprintf("Could not delete spreadsheet %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "deleted spreadsheet", map[string]interface{}{"id": state.ID.ValueString()})
}

// ImportState handles terraform import for existing Google Sheets spreadsheets.
// Usage: terraform import google_sheets_spreadsheet.example SPREADSHEET_ID
func (r *SpreadsheetResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
