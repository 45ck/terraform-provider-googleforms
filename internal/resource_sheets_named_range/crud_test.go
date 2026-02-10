// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsnamedrange

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
	r := &NamedRangeResource{}
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

func stateModel(t *testing.T, st tfsdk.State) NamedRangeResourceModel {
	t.Helper()
	var m NamedRangeResourceModel
	diags := st.Get(context.Background(), &m)
	if diags.HasError() {
		t.Fatalf("failed to decode state: %s", diags)
	}
	return m
}

func TestNamedRange_Create_SetsIDs(t *testing.T) {
	t.Parallel()

	mockSheets := &testutil.MockSheetsAPI{
		BatchUpdateFunc: func(_ context.Context, spreadsheetID string, req *sheets.BatchUpdateSpreadsheetRequest) (*sheets.BatchUpdateSpreadsheetResponse, error) {
			if spreadsheetID != "ss1" {
				t.Fatalf("unexpected spreadsheetID: %q", spreadsheetID)
			}
			if len(req.Requests) != 1 || req.Requests[0].AddNamedRange == nil {
				t.Fatalf("expected AddNamedRange request, got %#v", req.Requests)
			}
			nr := req.Requests[0].AddNamedRange.NamedRange
			if nr == nil || nr.Name != "MyRange" {
				t.Fatalf("unexpected named range: %#v", nr)
			}
			if nr.Range == nil || nr.Range.SheetId != 123 || nr.Range.StartRowIndex != 0 || nr.Range.EndRowIndex != 10 {
				t.Fatalf("unexpected grid range: %#v", nr.Range)
			}
			return &sheets.BatchUpdateSpreadsheetResponse{
				Replies: []*sheets.Response{
					{
						AddNamedRange: &sheets.AddNamedRangeResponse{
							NamedRange: &sheets.NamedRange{NamedRangeId: "nr-abc"},
						},
					},
				},
			}, nil
		},
	}

	r := &NamedRangeResource{client: &client.Client{Sheets: mockSheets}}

	rangeObj := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"sheet_id":           tftypes.Number,
				"start_row_index":    tftypes.Number,
				"end_row_index":      tftypes.Number,
				"start_column_index": tftypes.Number,
				"end_column_index":   tftypes.Number,
			},
		},
		map[string]tftypes.Value{
			"sheet_id":           tftypes.NewValue(tftypes.Number, big.NewFloat(123)),
			"start_row_index":    tftypes.NewValue(tftypes.Number, big.NewFloat(0)),
			"end_row_index":      tftypes.NewValue(tftypes.Number, big.NewFloat(10)),
			"start_column_index": tftypes.NewValue(tftypes.Number, big.NewFloat(0)),
			"end_column_index":   tftypes.NewValue(tftypes.Number, big.NewFloat(5)),
		},
	)

	plan := buildPlan(t, map[string]tftypes.Value{
		"spreadsheet_id": tftypes.NewValue(tftypes.String, "ss1"),
		"name":           tftypes.NewValue(tftypes.String, "MyRange"),
		"range":          rangeObj,
	})

	resp := &resource.CreateResponse{State: emptyState(t)}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %s", resp.Diagnostics)
	}

	got := stateModel(t, resp.State)
	if got.NamedRangeID.ValueString() != "nr-abc" {
		t.Fatalf("named_range_id=%q, want %q", got.NamedRangeID.ValueString(), "nr-abc")
	}
	if got.ID.ValueString() != "ss1#nr-abc" {
		t.Fatalf("id=%q, want %q", got.ID.ValueString(), "ss1#nr-abc")
	}
}
