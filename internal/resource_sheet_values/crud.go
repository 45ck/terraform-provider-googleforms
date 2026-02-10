// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetvalues

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sheets "google.golang.org/api/sheets/v4"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

func (r *SheetValuesResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SheetValuesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spreadsheetID := plan.SpreadsheetID.ValueString()
	rng := plan.Range.ValueString()

	vr, diags := planToValueRange(ctx, plan.Rows, rng)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateResp, err := r.client.Sheets.ValuesUpdate(ctx, spreadsheetID, rng, vr, plan.ValueInputOption.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Write Sheet Values Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(composeID(spreadsheetID, rng))
	plan.UpdatedRange = types.StringValue(updateResp.UpdatedRange)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	tflog.Debug(ctx, "wrote sheet values", map[string]interface{}{"id": plan.ID.ValueString()})
}

func (r *SheetValuesResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state SheetValuesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ReadBack.IsNull() || state.ReadBack.IsUnknown() || !state.ReadBack.ValueBool() {
		// Keep state as-is; drift is not detected in this mode.
		return
	}

	spreadsheetID := state.SpreadsheetID.ValueString()
	rng := state.Range.ValueString()

	vr, err := r.client.Sheets.ValuesGet(ctx, spreadsheetID, rng)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Sheet Values Failed", err.Error())
		return
	}

	rows, diags := valueRangeToRows(ctx, vr)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Rows = rows
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SheetValuesResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan SheetValuesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spreadsheetID := plan.SpreadsheetID.ValueString()
	rng := plan.Range.ValueString()

	vr, diags := planToValueRange(ctx, plan.Rows, rng)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateResp, err := r.client.Sheets.ValuesUpdate(ctx, spreadsheetID, rng, vr, plan.ValueInputOption.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Update Sheet Values Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(composeID(spreadsheetID, rng))
	plan.UpdatedRange = types.StringValue(updateResp.UpdatedRange)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SheetValuesResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state SheetValuesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Sheets.ValuesClear(ctx, state.SpreadsheetID.ValueString(), state.Range.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete Sheet Values Failed", err.Error())
		return
	}
}

func (r *SheetValuesResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Import format: spreadsheetID#A1_RANGE
	parts, diags := splitCompositeID(req.ID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("spreadsheet_id"), types.StringValue(parts.spreadsheetID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("range"), types.StringValue(parts.rng))...)
}

type compositeIDParts struct {
	spreadsheetID string
	rng           string
}

func splitCompositeID(id string) (compositeIDParts, diag.Diagnostics) {
	var diags diag.Diagnostics
	for i := 0; i < len(id); i++ {
		if id[i] == '#' {
			if i == 0 || i == len(id)-1 {
				break
			}
			return compositeIDParts{
				spreadsheetID: id[:i],
				rng:           id[i+1:],
			}, diags
		}
	}
	diags.AddError("Invalid Import ID", fmt.Sprintf("Expected import ID format spreadsheetID#A1_RANGE, got %q", id))
	return compositeIDParts{}, diags
}

func composeID(spreadsheetID, rng string) string {
	return fmt.Sprintf("%s#%s", spreadsheetID, rng)
}

func planToValueRange(ctx context.Context, rows types.List, rng string) (*sheets.ValueRange, diag.Diagnostics) {
	var diags diag.Diagnostics

	if rows.IsNull() || rows.IsUnknown() {
		diags.AddError("Invalid rows", "rows must be set")
		return nil, diags
	}

	var decoded []SheetValuesRowModel
	diags.Append(rows.ElementsAs(ctx, &decoded, false)...)
	if diags.HasError() {
		return nil, diags
	}

	values := make([][]interface{}, 0, len(decoded))
	for _, row := range decoded {
		var cells []string
		diags.Append(row.Cells.ElementsAs(ctx, &cells, false)...)
		if diags.HasError() {
			return nil, diags
		}
		line := make([]interface{}, 0, len(cells))
		for _, c := range cells {
			line = append(line, c)
		}
		values = append(values, line)
	}

	return &sheets.ValueRange{
		Range:  rng,
		Values: values,
	}, diags
}

func valueRangeToRows(_ context.Context, vr *sheets.ValueRange) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	rowType := types.ObjectType{AttrTypes: map[string]attr.Type{"cells": types.ListType{ElemType: types.StringType}}}
	out := make([]attr.Value, 0, len(vr.Values))

	for _, r := range vr.Values {
		cells := make([]attr.Value, 0, len(r))
		for _, c := range r {
			// The API can return numbers/bools; we normalize to strings.
			cells = append(cells, types.StringValue(fmt.Sprintf("%v", c)))
		}
		cellList, d := types.ListValue(types.StringType, cells)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(rowType), diags
		}

		obj, d := types.ObjectValue(rowType.AttrTypes, map[string]attr.Value{"cells": cellList})
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(rowType), diags
		}
		out = append(out, obj)
	}

	list, d := types.ListValue(rowType, out)
	diags.Append(d...)
	if diags.HasError() {
		return types.ListNull(rowType), diags
	}
	return list, diags
}
