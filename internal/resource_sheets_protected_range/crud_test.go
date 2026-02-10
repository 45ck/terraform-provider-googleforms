// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsprotectedrange

import (
	"context"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	sheets "google.golang.org/api/sheets/v4"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
	"github.com/45ck/terraform-provider-googleforms/internal/testutil"
)

func testSchemaResp() resource.SchemaResponse {
	var resp resource.SchemaResponse
	r := &ProtectedRangeResource{}
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

func buildPlan(t *testing.T, vals map[string]tftypes.Value) tfsdk.Plan {
	t.Helper()
	schemaResp := testSchemaResp()
	s := schemaResp.Schema
	tfType := s.Type().TerraformType(context.Background())
	objType, ok := tfType.(tftypes.Object)
	if !ok {
		t.Fatalf("expected tftypes.Object, got %T", tfType)
	}
	merged := make(map[string]tftypes.Value)
	for k, v := range objType.AttributeTypes {
		merged[k] = tftypes.NewValue(v, nil)
	}
	for k, v := range vals {
		merged[k] = v
	}
	return tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(objType, merged)}
}

func emptyState(t *testing.T) tfsdk.State {
	t.Helper()
	schemaResp := testSchemaResp()
	s := schemaResp.Schema
	tfType := s.Type().TerraformType(context.Background())
	objType, ok := tfType.(tftypes.Object)
	if !ok {
		t.Fatalf("expected tftypes.Object, got %T", tfType)
	}
	return tfsdk.State{Schema: s, Raw: tftypes.NewValue(objType, nil)}
}

func stateModel(t *testing.T, st tfsdk.State) ProtectedRangeResourceModel {
	t.Helper()
	var m ProtectedRangeResourceModel
	diags := st.Get(context.Background(), &m)
	if diags.HasError() {
		t.Fatalf("failed to decode state: %s", diags)
	}
	return m
}

func TestProtectedRange_Create_SetsIDs(t *testing.T) {
	t.Parallel()

	mockSheets := &testutil.MockSheetsAPI{
		BatchUpdateFunc: func(_ context.Context, spreadsheetID string, req *sheets.BatchUpdateSpreadsheetRequest) (*sheets.BatchUpdateSpreadsheetResponse, error) {
			if spreadsheetID != "ss1" {
				t.Fatalf("unexpected spreadsheetID: %q", spreadsheetID)
			}
			if len(req.Requests) != 1 || req.Requests[0].AddProtectedRange == nil {
				t.Fatalf("expected AddProtectedRange request")
			}
			pr := req.Requests[0].AddProtectedRange.ProtectedRange
			if pr == nil || pr.Range == nil || pr.Range.SheetId != 7 {
				t.Fatalf("unexpected protectedRange: %#v", pr)
			}
			return &sheets.BatchUpdateSpreadsheetResponse{
				Replies: []*sheets.Response{
					{
						AddProtectedRange: &sheets.AddProtectedRangeResponse{
							ProtectedRange: &sheets.ProtectedRange{ProtectedRangeId: 42},
						},
					},
				},
			}, nil
		},
	}

	r := &ProtectedRangeResource{client: &client.Client{Sheets: mockSheets}}

	rangeObj := tftypes.NewValue(
		tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"sheet_id": tftypes.Number, "start_row_index": tftypes.Number, "end_row_index": tftypes.Number, "start_column_index": tftypes.Number, "end_column_index": tftypes.Number,
		}},
		map[string]tftypes.Value{
			"sheet_id":           tftypes.NewValue(tftypes.Number, big.NewFloat(7)),
			"start_row_index":    tftypes.NewValue(tftypes.Number, big.NewFloat(0)),
			"end_row_index":      tftypes.NewValue(tftypes.Number, big.NewFloat(10)),
			"start_column_index": tftypes.NewValue(tftypes.Number, big.NewFloat(0)),
			"end_column_index":   tftypes.NewValue(tftypes.Number, big.NewFloat(5)),
		},
	)

	plan := buildPlan(t, map[string]tftypes.Value{
		"spreadsheet_id": tftypes.NewValue(tftypes.String, "ss1"),
		"description":    tftypes.NewValue(tftypes.String, "lock it"),
		"warning_only":   tftypes.NewValue(tftypes.Bool, false),
		"range":          rangeObj,
	})

	resp := &resource.CreateResponse{State: emptyState(t)}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %s", resp.Diagnostics)
	}

	got := stateModel(t, resp.State)
	if got.ProtectedRangeID.ValueInt64() != 42 {
		t.Fatalf("protected_range_id=%d, want %d", got.ProtectedRangeID.ValueInt64(), 42)
	}
	if got.ID.ValueString() != "ss1#42" {
		t.Fatalf("id=%q, want %q", got.ID.ValueString(), "ss1#42")
	}
}
