// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsprotectedrange

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sheets "google.golang.org/api/sheets/v4"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

func (r *ProtectedRangeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProtectedRangeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pr := &sheets.ProtectedRange{
		Description: plan.Description.ValueString(),
		WarningOnly: plan.WarningOnly.ValueBool(),
		Range:       toGridRange(plan.Range),
	}

	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddProtectedRange: &sheets.AddProtectedRangeRequest{ProtectedRange: pr},
			},
		},
	}

	apiResp, err := r.client.Sheets.BatchUpdate(ctx, plan.SpreadsheetID.ValueString(), batchReq)
	if err != nil {
		resp.Diagnostics.AddError("Add Protected Range Failed", err.Error())
		return
	}

	id, err := extractProtectedRangeID(apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Add Protected Range Failed", err.Error())
		return
	}

	plan.ProtectedRangeID = types.Int64Value(id)
	plan.ID = types.StringValue(composeID(plan.SpreadsheetID.ValueString(), id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	tflog.Debug(ctx, "created protected range", map[string]interface{}{"id": plan.ID.ValueString()})
}

func (r *ProtectedRangeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProtectedRangeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ss, err := r.client.Sheets.Get(ctx, state.SpreadsheetID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Spreadsheet Failed", err.Error())
		return
	}

	targetID := state.ProtectedRangeID.ValueInt64()
	found := false
	for _, sh := range ss.Sheets {
		if sh == nil {
			continue
		}
		for _, pr := range sh.ProtectedRanges {
			if pr != nil && pr.ProtectedRangeId == targetID {
				state.Description = types.StringValue(pr.Description)
				state.WarningOnly = types.BoolValue(pr.WarningOnly)
				state.Range = fromGridRange(pr.Range)
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ProtectedRangeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProtectedRangeResourceModel
	var state ProtectedRangeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	prID := state.ProtectedRangeID.ValueInt64()
	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				UpdateProtectedRange: &sheets.UpdateProtectedRangeRequest{
					ProtectedRange: &sheets.ProtectedRange{
						ProtectedRangeId: prID,
						Description:      plan.Description.ValueString(),
						WarningOnly:      plan.WarningOnly.ValueBool(),
						Range:            toGridRange(plan.Range),
					},
					Fields: "description,warningOnly,range",
				},
			},
		},
	}

	_, err := r.client.Sheets.BatchUpdate(ctx, state.SpreadsheetID.ValueString(), batchReq)
	if err != nil {
		resp.Diagnostics.AddError("Update Protected Range Failed", err.Error())
		return
	}

	plan.ProtectedRangeID = state.ProtectedRangeID
	plan.SpreadsheetID = state.SpreadsheetID
	plan.ID = types.StringValue(composeID(state.SpreadsheetID.ValueString(), prID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProtectedRangeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProtectedRangeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DeleteProtectedRange: &sheets.DeleteProtectedRangeRequest{
					ProtectedRangeId: state.ProtectedRangeID.ValueInt64(),
				},
			},
		},
	}

	_, err := r.client.Sheets.BatchUpdate(ctx, state.SpreadsheetID.ValueString(), batchReq)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete Protected Range Failed", err.Error())
		return
	}
}

func (r *ProtectedRangeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: spreadsheetID#protectedRangeId
	parts := strings.SplitN(req.ID, "#", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Expected import ID format spreadsheetID#protectedRangeId, got %q", req.ID))
		return
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("protectedRangeId must be an integer, got %q", parts[1]))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("spreadsheet_id"), types.StringValue(parts[0]))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("protected_range_id"), types.Int64Value(id))...)
}

func composeID(spreadsheetID string, protectedRangeID int64) string {
	return spreadsheetID + "#" + strconv.FormatInt(protectedRangeID, 10)
}

func toGridRange(m GridRangeModel) *sheets.GridRange {
	return &sheets.GridRange{
		SheetId:          m.SheetID.ValueInt64(),
		StartRowIndex:    m.StartRowIndex.ValueInt64(),
		EndRowIndex:      m.EndRowIndex.ValueInt64(),
		StartColumnIndex: m.StartColumnIndex.ValueInt64(),
		EndColumnIndex:   m.EndColumnIndex.ValueInt64(),
	}
}

func fromGridRange(gr *sheets.GridRange) GridRangeModel {
	return GridRangeModel{
		SheetID:          types.Int64Value(gr.SheetId),
		StartRowIndex:    types.Int64Value(gr.StartRowIndex),
		EndRowIndex:      types.Int64Value(gr.EndRowIndex),
		StartColumnIndex: types.Int64Value(gr.StartColumnIndex),
		EndColumnIndex:   types.Int64Value(gr.EndColumnIndex),
	}
}

func extractProtectedRangeID(resp *sheets.BatchUpdateSpreadsheetResponse) (int64, error) {
	if resp == nil || len(resp.Replies) == 0 || resp.Replies[0] == nil || resp.Replies[0].AddProtectedRange == nil || resp.Replies[0].AddProtectedRange.ProtectedRange == nil {
		return 0, fmt.Errorf("API did not return addProtectedRange reply")
	}
	id := resp.Replies[0].AddProtectedRange.ProtectedRange.ProtectedRangeId
	if id == 0 {
		return 0, fmt.Errorf("API returned empty protectedRangeId")
	}
	return id, nil
}
