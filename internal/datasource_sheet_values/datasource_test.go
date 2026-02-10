// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package datasourcesheetvalues

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	sheets "google.golang.org/api/sheets/v4"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
	"github.com/45ck/terraform-provider-googleforms/internal/testutil"
)

func testSchema(t *testing.T, ds datasource.DataSource) datasource.SchemaResponse {
	t.Helper()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func buildConfig(t *testing.T, schemaResp datasource.SchemaResponse, vals map[string]tftypes.Value) tfsdk.Config {
	t.Helper()
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

	return tfsdk.Config{
		Schema: s,
		Raw:    tftypes.NewValue(objType, merged),
	}
}

func TestSheetValuesDataSource_Metadata(t *testing.T) {
	t.Parallel()

	ds := NewSheetValuesDataSource()
	resp := &datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "googleforms"}, resp)

	if resp.TypeName != "googleforms_sheet_values" {
		t.Fatalf("unexpected type name: %q", resp.TypeName)
	}
}

func TestSheetValuesDataSource_ValueRangeToRows(t *testing.T) {
	t.Parallel()

	vr := &sheets.ValueRange{
		Values: [][]interface{}{
			{"a", 1},
			{true, nil},
		},
	}

	rows, diags := valueRangeToRows(vr)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags)
	}
	if rows.IsNull() || rows.IsUnknown() {
		t.Fatal("expected rows to be known")
	}
}

func TestSheetValuesDataSource_Read_SetsState(t *testing.T) {
	t.Parallel()

	mockSheets := &testutil.MockSheetsAPI{
		ValuesGetFunc: func(_ context.Context, spreadsheetID, rng string) (*sheets.ValueRange, error) {
			if spreadsheetID != "sid" {
				t.Fatalf("unexpected spreadsheetID: %q", spreadsheetID)
			}
			if rng != "Config!A1:B2" {
				t.Fatalf("unexpected range: %q", rng)
			}
			return &sheets.ValueRange{
				Range: rng,
				Values: [][]interface{}{
					{"a", "b"},
				},
			}, nil
		},
	}

	dsIface := NewSheetValuesDataSource()
	ds, ok := dsIface.(*SheetValuesDataSource)
	if !ok {
		t.Fatalf("expected *SheetValuesDataSource, got %T", dsIface)
	}
	ds.client = &client.Client{Sheets: mockSheets}

	schemaResp := testSchema(t, ds)
	cfg := buildConfig(t, schemaResp, map[string]tftypes.Value{
		"spreadsheet_id": tftypes.NewValue(tftypes.String, "sid"),
		"range":          tftypes.NewValue(tftypes.String, "Config!A1:B2"),
	})

	state := tfsdk.State{Schema: schemaResp.Schema}
	readResp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), datasource.ReadRequest{Config: cfg}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %s", readResp.Diagnostics)
	}
}
