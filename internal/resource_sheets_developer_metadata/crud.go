// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsdevelopermetadata

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

func (r *DeveloperMetadataResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DeveloperMetadataResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dm := &sheets.DeveloperMetadata{
		MetadataKey:   plan.MetadataKey.ValueString(),
		MetadataValue: plan.MetadataValue.ValueString(),
		Visibility:    plan.Visibility.ValueString(),
		Location:      expandLocation(plan.Location),
	}

	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				CreateDeveloperMetadata: &sheets.CreateDeveloperMetadataRequest{
					DeveloperMetadata: dm,
				},
			},
		},
	}

	apiResp, err := r.client.Sheets.BatchUpdate(ctx, plan.SpreadsheetID.ValueString(), batchReq)
	if err != nil {
		resp.Diagnostics.AddError("Add Developer Metadata Failed", err.Error())
		return
	}

	id, err := extractDeveloperMetadataID(apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Add Developer Metadata Failed", err.Error())
		return
	}

	plan.MetadataID = types.Int64Value(id)
	plan.ID = types.StringValue(composeID(plan.SpreadsheetID.ValueString(), id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	tflog.Debug(ctx, "created developer metadata", map[string]interface{}{"id": plan.ID.ValueString()})
}

func (r *DeveloperMetadataResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DeveloperMetadataResourceModel
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

	targetID := state.MetadataID.ValueInt64()

	found := false
	// Spreadsheet-level metadata.
	for _, md := range ss.DeveloperMetadata {
		if md != nil && md.MetadataId == targetID {
			state.MetadataKey = types.StringValue(md.MetadataKey)
			state.MetadataValue = types.StringValue(md.MetadataValue)
			state.Visibility = types.StringValue(md.Visibility)
			state.Location = flattenLocation(md.Location)
			found = true
			break
		}
	}

	// Sheet-level metadata can also exist.
	if !found {
		for _, sh := range ss.Sheets {
			if sh == nil {
				continue
			}
			for _, md := range sh.DeveloperMetadata {
				if md != nil && md.MetadataId == targetID {
					state.MetadataKey = types.StringValue(md.MetadataKey)
					state.MetadataValue = types.StringValue(md.MetadataValue)
					state.Visibility = types.StringValue(md.Visibility)
					state.Location = flattenLocation(md.Location)
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DeveloperMetadataResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DeveloperMetadataResourceModel
	var state DeveloperMetadataResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mdID := state.MetadataID.ValueInt64()

	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				UpdateDeveloperMetadata: &sheets.UpdateDeveloperMetadataRequest{
					DataFilters: []*sheets.DataFilter{
						{
							DeveloperMetadataLookup: &sheets.DeveloperMetadataLookup{
								MetadataId: mdID,
							},
						},
					},
					DeveloperMetadata: &sheets.DeveloperMetadata{
						MetadataId:    mdID,
						MetadataKey:   plan.MetadataKey.ValueString(),
						MetadataValue: plan.MetadataValue.ValueString(),
						Visibility:    plan.Visibility.ValueString(),
						Location:      expandLocation(plan.Location),
					},
					Fields: "metadataKey,metadataValue,visibility,location",
				},
			},
		},
	}

	_, err := r.client.Sheets.BatchUpdate(ctx, state.SpreadsheetID.ValueString(), batchReq)
	if err != nil {
		resp.Diagnostics.AddError("Update Developer Metadata Failed", err.Error())
		return
	}

	plan.MetadataID = state.MetadataID
	plan.SpreadsheetID = state.SpreadsheetID
	plan.ID = types.StringValue(composeID(state.SpreadsheetID.ValueString(), mdID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeveloperMetadataResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DeveloperMetadataResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DeleteDeveloperMetadata: &sheets.DeleteDeveloperMetadataRequest{
					DataFilter: &sheets.DataFilter{
						DeveloperMetadataLookup: &sheets.DeveloperMetadataLookup{
							MetadataId: state.MetadataID.ValueInt64(),
						},
					},
				},
			},
		},
	}

	_, err := r.client.Sheets.BatchUpdate(ctx, state.SpreadsheetID.ValueString(), batchReq)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete Developer Metadata Failed", err.Error())
		return
	}
}

func (r *DeveloperMetadataResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: spreadsheetID#metadataId
	parts := strings.SplitN(req.ID, "#", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Expected import ID format spreadsheetID#metadataId, got %q", req.ID))
		return
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("metadataId must be an integer, got %q", parts[1]))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("spreadsheet_id"), types.StringValue(parts[0]))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("metadata_id"), types.Int64Value(id))...)
}

func composeID(spreadsheetID string, metadataID int64) string {
	return spreadsheetID + "#" + strconv.FormatInt(metadataID, 10)
}

func expandLocation(loc *MetadataLocationModel) *sheets.DeveloperMetadataLocation {
	if loc == nil {
		return &sheets.DeveloperMetadataLocation{Spreadsheet: true}
	}

	if !loc.SheetID.IsNull() && !loc.SheetID.IsUnknown() {
		return &sheets.DeveloperMetadataLocation{SheetId: loc.SheetID.ValueInt64()}
	}

	return &sheets.DeveloperMetadataLocation{Spreadsheet: true}
}

func flattenLocation(loc *sheets.DeveloperMetadataLocation) *MetadataLocationModel {
	if loc == nil {
		return nil
	}
	if loc.SheetId != 0 {
		return &MetadataLocationModel{SheetID: types.Int64Value(loc.SheetId)}
	}
	return nil
}

func extractDeveloperMetadataID(resp *sheets.BatchUpdateSpreadsheetResponse) (int64, error) {
	if resp == nil || len(resp.Replies) == 0 || resp.Replies[0] == nil || resp.Replies[0].CreateDeveloperMetadata == nil || resp.Replies[0].CreateDeveloperMetadata.DeveloperMetadata == nil {
		return 0, fmt.Errorf("API did not return createDeveloperMetadata reply")
	}
	id := resp.Replies[0].CreateDeveloperMetadata.DeveloperMetadata.MetadataId
	if id == 0 {
		return 0, fmt.Errorf("API returned empty metadataId")
	}
	return id, nil
}
