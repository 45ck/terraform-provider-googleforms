// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheet

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sheets "google.golang.org/api/sheets/v4"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

// Create adds a new sheet to the specified spreadsheet.
func (r *SheetResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SheetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spreadsheetID := plan.SpreadsheetID.ValueString()

	tflog.Debug(ctx, "creating sheet", map[string]interface{}{
		"spreadsheet_id": spreadsheetID,
		"title":          plan.Title.ValueString(),
	})

	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{
						Title: plan.Title.ValueString(),
						GridProperties: &sheets.GridProperties{
							RowCount:    plan.RowCount.ValueInt64(),
							ColumnCount: plan.ColumnCount.ValueInt64(),
						},
					},
				},
			},
		},
	}

	batchResp, err := r.client.Sheets.BatchUpdate(ctx, spreadsheetID, batchReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Sheet",
			fmt.Sprintf("Could not add sheet to spreadsheet %s: %s", spreadsheetID, err),
		)
		return
	}

	// Extract the sheet ID from the AddSheet reply.
	addedProps := batchResp.Replies[0].AddSheet.Properties
	sheetID := addedProps.SheetId

	tflog.Info(ctx, "created sheet", map[string]interface{}{
		"spreadsheet_id": spreadsheetID,
		"sheet_id":       sheetID,
	})

	// Build composite ID and save partial state immediately.
	plan.ID = types.StringValue(fmt.Sprintf("%s#%d", spreadsheetID, sheetID))
	plan.SheetID = types.Int64Value(sheetID)
	plan.Index = types.Int64Value(int64(addedProps.Index))
	plan.RowCount = types.Int64Value(addedProps.GridProperties.RowCount)
	plan.ColumnCount = types.Int64Value(addedProps.GridProperties.ColumnCount)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read fetches the current state of a sheet from the Sheets API.
func (r *SheetResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state SheetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spreadsheetID, sheetID, diags := parseSheetID(state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "reading sheet", map[string]interface{}{
		"spreadsheet_id": spreadsheetID,
		"sheet_id":       sheetID,
	})

	// Fetch the entire spreadsheet to find our sheet.
	spreadsheet, err := r.client.Sheets.Get(ctx, spreadsheetID)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "spreadsheet not found, removing sheet from state", map[string]interface{}{
				"spreadsheet_id": spreadsheetID,
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Sheet",
			fmt.Sprintf("Could not read spreadsheet %s: %s", spreadsheetID, err),
		)
		return
	}

	// Find the sheet by matching SheetId.
	var found *sheets.SheetProperties
	for _, s := range spreadsheet.Sheets {
		if s.Properties.SheetId == sheetID {
			found = s.Properties
			break
		}
	}

	if found == nil {
		tflog.Warn(ctx, "sheet not found in spreadsheet, removing from state", map[string]interface{}{
			"spreadsheet_id": spreadsheetID,
			"sheet_id":       sheetID,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state from the API response.
	state.SpreadsheetID = types.StringValue(spreadsheetID)
	state.Title = types.StringValue(found.Title)
	state.SheetID = types.Int64Value(found.SheetId)
	state.Index = types.Int64Value(int64(found.Index))
	state.RowCount = types.Int64Value(found.GridProperties.RowCount)
	state.ColumnCount = types.Int64Value(found.GridProperties.ColumnCount)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies the sheet's title and grid properties.
func (r *SheetResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan SheetResourceModel
	var state SheetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spreadsheetID, sheetID, diags := parseSheetID(state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "updating sheet", map[string]interface{}{
		"spreadsheet_id": spreadsheetID,
		"sheet_id":       sheetID,
	})

	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
					Properties: &sheets.SheetProperties{
						SheetId: sheetID,
						Title:   plan.Title.ValueString(),
						GridProperties: &sheets.GridProperties{
							RowCount:    plan.RowCount.ValueInt64(),
							ColumnCount: plan.ColumnCount.ValueInt64(),
						},
					},
					Fields: "title,gridProperties.rowCount,gridProperties.columnCount",
				},
			},
		},
	}

	_, err := r.client.Sheets.BatchUpdate(ctx, spreadsheetID, batchReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Sheet",
			fmt.Sprintf("Could not update sheet %d in spreadsheet %s: %s", sheetID, spreadsheetID, err),
		)
		return
	}

	// Read back the updated state.
	spreadsheet, err := r.client.Sheets.Get(ctx, spreadsheetID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Sheet After Update",
			fmt.Sprintf("Could not read spreadsheet %s after update: %s", spreadsheetID, err),
		)
		return
	}

	var found *sheets.SheetProperties
	for _, s := range spreadsheet.Sheets {
		if s.Properties.SheetId == sheetID {
			found = s.Properties
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError(
			"Error Reading Sheet After Update",
			fmt.Sprintf("Sheet %d not found in spreadsheet %s after update", sheetID, spreadsheetID),
		)
		return
	}

	plan.ID = state.ID
	plan.SpreadsheetID = types.StringValue(spreadsheetID)
	plan.SheetID = types.Int64Value(found.SheetId)
	plan.Index = types.Int64Value(int64(found.Index))
	plan.RowCount = types.Int64Value(found.GridProperties.RowCount)
	plan.ColumnCount = types.Int64Value(found.GridProperties.ColumnCount)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a sheet from the spreadsheet.
func (r *SheetResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state SheetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spreadsheetID, sheetID, diags := parseSheetID(state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting sheet", map[string]interface{}{
		"spreadsheet_id": spreadsheetID,
		"sheet_id":       sheetID,
	})

	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DeleteSheet: &sheets.DeleteSheetRequest{
					SheetId: sheetID,
				},
			},
		},
	}

	_, err := r.client.Sheets.BatchUpdate(ctx, spreadsheetID, batchReq)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "spreadsheet or sheet already deleted", map[string]interface{}{
				"spreadsheet_id": spreadsheetID,
				"sheet_id":       sheetID,
			})
			return
		}

		resp.Diagnostics.AddError(
			"Error Deleting Sheet",
			fmt.Sprintf("Could not delete sheet %d from spreadsheet %s: %s", sheetID, spreadsheetID, err),
		)
		return
	}

	tflog.Info(ctx, "deleted sheet", map[string]interface{}{
		"spreadsheet_id": spreadsheetID,
		"sheet_id":       sheetID,
	})

	// State is automatically removed by the framework after Delete returns
	// without errors.
}

// ImportState handles terraform import for existing sheets.
// Usage: terraform import google_forms_sheet.example SPREADSHEET_ID#SHEET_ID
func (r *SheetResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// parseSheetID splits a composite ID "spreadsheetID#sheetID" into its components.
func parseSheetID(id string) (string, int64, diag.Diagnostics) {
	var diags diag.Diagnostics

	parts := strings.SplitN(id, "#", 2)
	if len(parts) != 2 {
		diags.AddError(
			"Invalid Sheet ID Format",
			fmt.Sprintf("Expected format 'spreadsheetID#sheetID', got: %s", id),
		)
		return "", 0, diags
	}

	sheetID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		diags.AddError(
			"Invalid Sheet ID",
			fmt.Sprintf("Could not parse sheet ID as integer from '%s': %s", parts[1], err),
		)
		return "", 0, diags
	}

	return parts[0], sheetID, diags
}
