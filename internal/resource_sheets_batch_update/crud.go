// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsbatchupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sheets "google.golang.org/api/sheets/v4"
)

func (r *SheetsBatchUpdateResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SheetsBatchUpdateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	batchReq, err := decodeBatchUpdateRequest(plan.RequestsJSON.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid requests_json", err.Error())
		return
	}
	batchReq.IncludeSpreadsheetInResponse = plan.IncludeSpreadsheetInResponse.ValueBool()

	apiResp, err := r.client.Sheets.BatchUpdate(ctx, plan.SpreadsheetID.ValueString(), batchReq)
	if err != nil {
		resp.Diagnostics.AddError("Sheets batchUpdate failed", err.Error())
		return
	}

	plan.ID = types.StringValue(hashID(plan.SpreadsheetID.ValueString(), plan.RequestsJSON.ValueString()))

	if plan.StoreResponseJSON.ValueBool() {
		j, jerr := json.Marshal(apiResp)
		if jerr != nil {
			resp.Diagnostics.AddError("Failed to serialize response", jerr.Error())
			return
		}
		plan.ResponseJSON = types.StringValue(string(j))
	} else {
		plan.ResponseJSON = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	tflog.Debug(ctx, "applied sheets batchUpdate", map[string]interface{}{"id": plan.ID.ValueString()})
}

func (r *SheetsBatchUpdateResource) Read(
	_ context.Context,
	_ resource.ReadRequest,
	_ *resource.ReadResponse,
) {
	// No read: this resource is an imperative escape hatch.
}

func (r *SheetsBatchUpdateResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// Re-apply on changes.
	var plan SheetsBatchUpdateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	batchReq, err := decodeBatchUpdateRequest(plan.RequestsJSON.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid requests_json", err.Error())
		return
	}
	batchReq.IncludeSpreadsheetInResponse = plan.IncludeSpreadsheetInResponse.ValueBool()

	apiResp, err := r.client.Sheets.BatchUpdate(ctx, plan.SpreadsheetID.ValueString(), batchReq)
	if err != nil {
		resp.Diagnostics.AddError("Sheets batchUpdate failed", err.Error())
		return
	}

	plan.ID = types.StringValue(hashID(plan.SpreadsheetID.ValueString(), plan.RequestsJSON.ValueString()))

	if plan.StoreResponseJSON.ValueBool() {
		j, jerr := json.Marshal(apiResp)
		if jerr != nil {
			resp.Diagnostics.AddError("Failed to serialize response", jerr.Error())
			return
		}
		plan.ResponseJSON = types.StringValue(string(j))
	} else {
		plan.ResponseJSON = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SheetsBatchUpdateResource) Delete(
	_ context.Context,
	_ resource.DeleteRequest,
	_ *resource.DeleteResponse,
) {
	// No-op: Sheets batchUpdate requests are not reversible in general.
}

func decodeBatchUpdateRequest(s string) (*sheets.BatchUpdateSpreadsheetRequest, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil, fmt.Errorf("requests_json must not be empty")
	}

	// Accept either:
	// 1) array of Request objects: [ { ... }, ... ]
	// 2) full BatchUpdateSpreadsheetRequest: { "requests": [ ... ], ... }
	if strings.HasPrefix(trimmed, "[") {
		var reqs []*sheets.Request
		if err := json.Unmarshal([]byte(trimmed), &reqs); err != nil {
			return nil, fmt.Errorf("parsing requests array: %w", err)
		}
		return &sheets.BatchUpdateSpreadsheetRequest{Requests: reqs}, nil
	}

	var batch sheets.BatchUpdateSpreadsheetRequest
	if err := json.Unmarshal([]byte(trimmed), &batch); err != nil {
		return nil, fmt.Errorf("parsing batch update request object: %w", err)
	}
	if len(batch.Requests) == 0 {
		return nil, fmt.Errorf("batch update request must include at least one request")
	}
	return &batch, nil
}

func hashID(spreadsheetID, requestsJSON string) string {
	sum := sha256.Sum256([]byte(spreadsheetID + "\n" + requestsJSON))
	return hex.EncodeToString(sum[:])
}
