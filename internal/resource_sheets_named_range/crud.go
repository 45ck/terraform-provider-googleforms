// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsnamedrange

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sheets "google.golang.org/api/sheets/v4"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

func (r *NamedRangeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NamedRangeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	grid := toGridRange(plan.Range)
	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddNamedRange: &sheets.AddNamedRangeRequest{
					NamedRange: &sheets.NamedRange{
						Name:  plan.Name.ValueString(),
						Range: grid,
					},
				},
			},
		},
	}

	apiResp, err := r.client.Sheets.BatchUpdate(ctx, plan.SpreadsheetID.ValueString(), batchReq)
	if err != nil {
		resp.Diagnostics.AddError("Add Named Range Failed", err.Error())
		return
	}

	nrID, err := extractNamedRangeID(apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Add Named Range Failed", err.Error())
		return
	}

	plan.NamedRangeID = types.StringValue(nrID)
	plan.ID = types.StringValue(composeID(plan.SpreadsheetID.ValueString(), nrID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	tflog.Debug(ctx, "created named range", map[string]interface{}{"id": plan.ID.ValueString()})
}

func (r *NamedRangeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NamedRangeResourceModel
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

	found := false
	for _, nr := range ss.NamedRanges {
		if nr != nil && nr.NamedRangeId == state.NamedRangeID.ValueString() {
			state.Name = types.StringValue(nr.Name)
			state.Range = fromGridRange(nr.Range)
			found = true
			break
		}
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NamedRangeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NamedRangeResourceModel
	var state NamedRangeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nrID := state.NamedRangeID.ValueString()
	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				UpdateNamedRange: &sheets.UpdateNamedRangeRequest{
					NamedRange: &sheets.NamedRange{
						NamedRangeId: nrID,
						Name:         plan.Name.ValueString(),
						Range:        toGridRange(plan.Range),
					},
					Fields: "name,range",
				},
			},
		},
	}

	_, err := r.client.Sheets.BatchUpdate(ctx, state.SpreadsheetID.ValueString(), batchReq)
	if err != nil {
		resp.Diagnostics.AddError("Update Named Range Failed", err.Error())
		return
	}

	plan.NamedRangeID = state.NamedRangeID
	plan.SpreadsheetID = state.SpreadsheetID
	plan.ID = types.StringValue(composeID(state.SpreadsheetID.ValueString(), nrID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NamedRangeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NamedRangeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DeleteNamedRange: &sheets.DeleteNamedRangeRequest{
					NamedRangeId: state.NamedRangeID.ValueString(),
				},
			},
		},
	}

	_, err := r.client.Sheets.BatchUpdate(ctx, state.SpreadsheetID.ValueString(), batchReq)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete Named Range Failed", err.Error())
		return
	}
}

func (r *NamedRangeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: spreadsheetID#namedRangeId
	parts := strings.SplitN(req.ID, "#", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Expected import ID format spreadsheetID#namedRangeId, got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("spreadsheet_id"), types.StringValue(parts[0]))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("named_range_id"), types.StringValue(parts[1]))...)
}

func composeID(spreadsheetID, namedRangeID string) string {
	return spreadsheetID + "#" + namedRangeID
}

func toGridRange(m GridRangeModel) *sheets.GridRange {
	gr := &sheets.GridRange{
		SheetId: m.SheetID.ValueInt64(),
	}
	gr.StartRowIndex = m.StartRowIndex.ValueInt64()
	gr.EndRowIndex = m.EndRowIndex.ValueInt64()
	gr.StartColumnIndex = m.StartColumnIndex.ValueInt64()
	gr.EndColumnIndex = m.EndColumnIndex.ValueInt64()
	return gr
}

func fromGridRange(gr *sheets.GridRange) GridRangeModel {
	out := GridRangeModel{
		SheetID: types.Int64Value(gr.SheetId),
	}
	out.StartRowIndex = types.Int64Value(gr.StartRowIndex)
	out.EndRowIndex = types.Int64Value(gr.EndRowIndex)
	out.StartColumnIndex = types.Int64Value(gr.StartColumnIndex)
	out.EndColumnIndex = types.Int64Value(gr.EndColumnIndex)
	return out
}

func extractNamedRangeID(resp *sheets.BatchUpdateSpreadsheetResponse) (string, error) {
	if resp == nil || len(resp.Replies) == 0 || resp.Replies[0] == nil || resp.Replies[0].AddNamedRange == nil || resp.Replies[0].AddNamedRange.NamedRange == nil {
		return "", fmt.Errorf("API did not return addNamedRange reply")
	}
	id := strings.TrimSpace(resp.Replies[0].AddNamedRange.NamedRange.NamedRangeId)
	if id == "" {
		return "", fmt.Errorf("API returned empty namedRangeId")
	}
	return id, nil
}
