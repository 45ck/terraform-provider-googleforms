// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsconditionalformatrule

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sheets "google.golang.org/api/sheets/v4"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

func (r *ConditionalFormatRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ConditionalFormatRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := decodeRule(plan.RuleJSON.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid rule_json", err.Error())
		return
	}

	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddConditionalFormatRule: &sheets.AddConditionalFormatRuleRequest{
					Index: plan.Index.ValueInt64(),
					Rule:  rule,
				},
			},
		},
	}

	_, err = r.client.Sheets.BatchUpdate(ctx, plan.SpreadsheetID.ValueString(), batchReq)
	if err != nil {
		resp.Diagnostics.AddError("Add Conditional Format Rule Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(composeID(plan.SpreadsheetID.ValueString(), plan.SheetID.ValueInt64(), plan.Index.ValueInt64()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConditionalFormatRuleResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
	// No read: rules are addressed by index and frequently modified by users.
}

func (r *ConditionalFormatRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ConditionalFormatRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := decodeRule(plan.RuleJSON.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid rule_json", err.Error())
		return
	}

	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				UpdateConditionalFormatRule: &sheets.UpdateConditionalFormatRuleRequest{
					Index:   plan.Index.ValueInt64(),
					Rule:    rule,
					SheetId: plan.SheetID.ValueInt64(),
				},
			},
		},
	}

	_, err = r.client.Sheets.BatchUpdate(ctx, plan.SpreadsheetID.ValueString(), batchReq)
	if err != nil {
		resp.Diagnostics.AddError("Update Conditional Format Rule Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(composeID(plan.SpreadsheetID.ValueString(), plan.SheetID.ValueInt64(), plan.Index.ValueInt64()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConditionalFormatRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ConditionalFormatRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DeleteConditionalFormatRule: &sheets.DeleteConditionalFormatRuleRequest{
					Index:   state.Index.ValueInt64(),
					SheetId: state.SheetID.ValueInt64(),
				},
			},
		},
	}

	_, err := r.client.Sheets.BatchUpdate(ctx, state.SpreadsheetID.ValueString(), batchReq)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete Conditional Format Rule Failed", err.Error())
		return
	}
}

func (r *ConditionalFormatRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID: spreadsheetID#sheetId#index
	parts := strings.SplitN(req.ID, "#", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Expected import ID format spreadsheetID#sheetId#index, got %q", req.ID))
		return
	}
	sheetID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("sheetId must be an integer, got %q", parts[1]))
		return
	}
	idx, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("index must be an integer, got %q", parts[2]))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("spreadsheet_id"), types.StringValue(parts[0]))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("sheet_id"), types.Int64Value(sheetID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("index"), types.Int64Value(idx))...)
}

func decodeRule(s string) (*sheets.ConditionalFormatRule, error) {
	var rule sheets.ConditionalFormatRule
	if err := json.Unmarshal([]byte(strings.TrimSpace(s)), &rule); err != nil {
		return nil, fmt.Errorf("parsing ConditionalFormatRule JSON: %w", err)
	}
	return &rule, nil
}

func composeID(spreadsheetID string, sheetID, index int64) string {
	return spreadsheetID + "#" + strconv.FormatInt(sheetID, 10) + "#" + strconv.FormatInt(index, 10)
}
