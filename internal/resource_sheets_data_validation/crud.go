// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsdatavalidation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sheets "google.golang.org/api/sheets/v4"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

func (r *DataValidationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DataValidationResourceModel
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
				SetDataValidation: &sheets.SetDataValidationRequest{
					Range: toGridRange(plan.Range),
					Rule:  rule,
				},
			},
		},
	}

	_, err = r.client.Sheets.BatchUpdate(ctx, plan.SpreadsheetID.ValueString(), batchReq)
	if err != nil {
		resp.Diagnostics.AddError("Set Data Validation Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(deterministicID(plan.SpreadsheetID.ValueString(), plan.Range))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DataValidationResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
	// No read: retrieving validation rules requires grid data and is expensive.
}

func (r *DataValidationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DataValidationResourceModel
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
				SetDataValidation: &sheets.SetDataValidationRequest{
					Range: toGridRange(plan.Range),
					Rule:  rule,
				},
			},
		},
	}

	_, err = r.client.Sheets.BatchUpdate(ctx, plan.SpreadsheetID.ValueString(), batchReq)
	if err != nil {
		resp.Diagnostics.AddError("Set Data Validation Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(deterministicID(plan.SpreadsheetID.ValueString(), plan.Range))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DataValidationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DataValidationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Clear validation by omitting Rule (per API contract).
	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				SetDataValidation: &sheets.SetDataValidationRequest{
					Range: toGridRange(state.Range),
					Rule:  nil,
				},
			},
		},
	}

	_, err := r.client.Sheets.BatchUpdate(ctx, state.SpreadsheetID.ValueString(), batchReq)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Clear Data Validation Failed", err.Error())
		return
	}
}

func (r *DataValidationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by id only; user must set spreadsheet_id/range/rule_json after import.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
}

func decodeRule(s string) (*sheets.DataValidationRule, error) {
	var rule sheets.DataValidationRule
	if err := json.Unmarshal([]byte(strings.TrimSpace(s)), &rule); err != nil {
		return nil, fmt.Errorf("parsing DataValidationRule JSON: %w", err)
	}
	return &rule, nil
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

func deterministicID(spreadsheetID string, gr GridRangeModel) string {
	raw := fmt.Sprintf("%s\n%d:%d\n%d:%d\n%d",
		spreadsheetID,
		gr.StartRowIndex.ValueInt64(), gr.EndRowIndex.ValueInt64(),
		gr.StartColumnIndex.ValueInt64(), gr.EndColumnIndex.ValueInt64(),
		gr.SheetID.ValueInt64(),
	)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
