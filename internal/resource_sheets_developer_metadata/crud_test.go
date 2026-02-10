// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsdevelopermetadata

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
	r := &DeveloperMetadataResource{}
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

func stateModel(t *testing.T, st tfsdk.State) DeveloperMetadataResourceModel {
	t.Helper()
	var m DeveloperMetadataResourceModel
	diags := st.Get(context.Background(), &m)
	if diags.HasError() {
		t.Fatalf("failed to decode state: %s", diags)
	}
	return m
}

func TestDeveloperMetadata_Create_SetsIDs(t *testing.T) {
	t.Parallel()

	mockSheets := &testutil.MockSheetsAPI{
		BatchUpdateFunc: func(_ context.Context, spreadsheetID string, req *sheets.BatchUpdateSpreadsheetRequest) (*sheets.BatchUpdateSpreadsheetResponse, error) {
			if spreadsheetID != "ss1" {
				t.Fatalf("unexpected spreadsheetID: %q", spreadsheetID)
			}
			if len(req.Requests) != 1 || req.Requests[0].CreateDeveloperMetadata == nil {
				t.Fatalf("expected CreateDeveloperMetadata request")
			}
			dm := req.Requests[0].CreateDeveloperMetadata.DeveloperMetadata
			if dm == nil || dm.MetadataKey != "k" || dm.Visibility != "DOCUMENT" {
				t.Fatalf("unexpected developer metadata: %#v", dm)
			}
			if dm.Location == nil || dm.Location.SheetId != 99 {
				t.Fatalf("expected sheet-scoped location, got %#v", dm.Location)
			}
			return &sheets.BatchUpdateSpreadsheetResponse{
				Replies: []*sheets.Response{
					{
						CreateDeveloperMetadata: &sheets.CreateDeveloperMetadataResponse{
							DeveloperMetadata: &sheets.DeveloperMetadata{MetadataId: 777},
						},
					},
				},
			}, nil
		},
	}

	r := &DeveloperMetadataResource{client: &client.Client{Sheets: mockSheets}}

	locObj := tftypes.NewValue(
		tftypes.Object{AttributeTypes: map[string]tftypes.Type{"sheet_id": tftypes.Number}},
		map[string]tftypes.Value{"sheet_id": tftypes.NewValue(tftypes.Number, big.NewFloat(99))},
	)

	plan := buildPlan(t, map[string]tftypes.Value{
		"spreadsheet_id": tftypes.NewValue(tftypes.String, "ss1"),
		"metadata_key":   tftypes.NewValue(tftypes.String, "k"),
		"metadata_value": tftypes.NewValue(tftypes.String, "v"),
		"visibility":     tftypes.NewValue(tftypes.String, "DOCUMENT"),
		"location":       locObj,
	})

	resp := &resource.CreateResponse{State: emptyState(t)}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %s", resp.Diagnostics)
	}

	got := stateModel(t, resp.State)
	if got.MetadataID.ValueInt64() != 777 {
		t.Fatalf("metadata_id=%d, want %d", got.MetadataID.ValueInt64(), 777)
	}
	if got.ID.ValueString() != "ss1#777" {
		t.Fatalf("id=%q, want %q", got.ID.ValueString(), "ss1#777")
	}
}
