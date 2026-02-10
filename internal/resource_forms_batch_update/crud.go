// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceformsbatchupdate

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
	forms "google.golang.org/api/forms/v1"
)

func (r *FormsBatchUpdateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FormsBatchUpdateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	batchReq, err := decodeBatchUpdateRequest(plan.RequestsJSON.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid requests_json", err.Error())
		return
	}

	batchReq.IncludeFormInResponse = plan.IncludeFormInResponse.ValueBool()
	if !plan.RequiredRevisionID.IsNull() && !plan.RequiredRevisionID.IsUnknown() && strings.TrimSpace(plan.RequiredRevisionID.ValueString()) != "" {
		batchReq.WriteControl = &forms.WriteControl{RequiredRevisionId: plan.RequiredRevisionID.ValueString()}
	}

	apiResp, err := r.client.Forms.BatchUpdate(ctx, plan.FormID.ValueString(), batchReq)
	if err != nil {
		resp.Diagnostics.AddError("Forms batchUpdate failed", err.Error())
		return
	}

	plan.ID = types.StringValue(hashID(plan.FormID.ValueString(), plan.RequestsJSON.ValueString(), plan.RequiredRevisionID.ValueString(), plan.IncludeFormInResponse.ValueBool()))

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
	tflog.Debug(ctx, "applied forms batchUpdate", map[string]interface{}{"id": plan.ID.ValueString()})
}

func (r *FormsBatchUpdateResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
	// No read: this resource is an imperative escape hatch.
}

func (r *FormsBatchUpdateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FormsBatchUpdateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	batchReq, err := decodeBatchUpdateRequest(plan.RequestsJSON.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid requests_json", err.Error())
		return
	}

	batchReq.IncludeFormInResponse = plan.IncludeFormInResponse.ValueBool()
	if !plan.RequiredRevisionID.IsNull() && !plan.RequiredRevisionID.IsUnknown() && strings.TrimSpace(plan.RequiredRevisionID.ValueString()) != "" {
		batchReq.WriteControl = &forms.WriteControl{RequiredRevisionId: plan.RequiredRevisionID.ValueString()}
	}

	apiResp, err := r.client.Forms.BatchUpdate(ctx, plan.FormID.ValueString(), batchReq)
	if err != nil {
		resp.Diagnostics.AddError("Forms batchUpdate failed", err.Error())
		return
	}

	plan.ID = types.StringValue(hashID(plan.FormID.ValueString(), plan.RequestsJSON.ValueString(), plan.RequiredRevisionID.ValueString(), plan.IncludeFormInResponse.ValueBool()))

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

func (r *FormsBatchUpdateResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No-op: Forms batchUpdate requests are not reversible in general.
}

func decodeBatchUpdateRequest(s string) (*forms.BatchUpdateFormRequest, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil, fmt.Errorf("requests_json must not be empty")
	}

	// Accept either:
	// 1) array of Request objects: [ { ... }, ... ]
	// 2) full BatchUpdateFormRequest: { "requests": [ ... ], ... }
	if strings.HasPrefix(trimmed, "[") {
		var reqs []*forms.Request
		if err := json.Unmarshal([]byte(trimmed), &reqs); err != nil {
			return nil, fmt.Errorf("parsing requests array: %w", err)
		}
		if len(reqs) == 0 {
			return nil, fmt.Errorf("requests array must include at least one request")
		}
		return &forms.BatchUpdateFormRequest{Requests: reqs}, nil
	}

	var batch forms.BatchUpdateFormRequest
	if err := json.Unmarshal([]byte(trimmed), &batch); err != nil {
		return nil, fmt.Errorf("parsing batch update request object: %w", err)
	}
	if len(batch.Requests) == 0 {
		return nil, fmt.Errorf("batch update request must include at least one request")
	}
	return &batch, nil
}

func hashID(formID, requestsJSON, requiredRevisionID string, includeFormInResponse bool) string {
	sum := sha256.Sum256([]byte(formID + "\n" + requestsJSON + "\n" + requiredRevisionID + "\n" + fmt.Sprintf("%t", includeFormInResponse)))
	return hex.EncodeToString(sum[:])
}
