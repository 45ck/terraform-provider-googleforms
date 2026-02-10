// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package datasourcespreadsheet

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
	"github.com/45ck/terraform-provider-googleforms/internal/testutil"
	sheets "google.golang.org/api/sheets/v4"
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

func TestSpreadsheetDataSource_Metadata(t *testing.T) {
	t.Parallel()

	ds := NewSpreadsheetDataSource()
	resp := &datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "googleforms"}, resp)

	if resp.TypeName != "googleforms_spreadsheet" {
		t.Fatalf("unexpected type name: %q", resp.TypeName)
	}
}

func TestSpreadsheetDataSource_Configure_WrongType(t *testing.T) {
	t.Parallel()

	dsIface := NewSpreadsheetDataSource()
	ds, ok := dsIface.(*SpreadsheetDataSource)
	if !ok {
		t.Fatalf("expected *SpreadsheetDataSource, got %T", dsIface)
	}
	resp := &datasource.ConfigureResponse{}
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "not-a-client"}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostics")
	}
}

func TestSpreadsheetDataSource_Read_SetsState(t *testing.T) {
	t.Parallel()

	mockSheets := &testutil.MockSheetsAPI{
		GetFunc: func(_ context.Context, spreadsheetID string) (*sheets.Spreadsheet, error) {
			return &sheets.Spreadsheet{
				SpreadsheetId:  spreadsheetID,
				SpreadsheetUrl: "https://docs.google.com/spreadsheets/d/" + spreadsheetID + "/edit",
				Properties: &sheets.SpreadsheetProperties{
					Title:    "T",
					Locale:   "en_US",
					TimeZone: "UTC",
				},
			}, nil
		},
	}

	c := &client.Client{Sheets: mockSheets}

	dsIface := NewSpreadsheetDataSource()
	ds, ok := dsIface.(*SpreadsheetDataSource)
	if !ok {
		t.Fatalf("expected *SpreadsheetDataSource, got %T", dsIface)
	}
	ds.client = c

	schemaResp := testSchema(t, ds)
	cfg := buildConfig(t, schemaResp, map[string]tftypes.Value{
		"id": tftypes.NewValue(tftypes.String, "sid"),
	})

	state := tfsdk.State{Schema: schemaResp.Schema}
	readResp := &datasource.ReadResponse{State: state}

	ds.Read(context.Background(), datasource.ReadRequest{Config: cfg}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %s", readResp.Diagnostics)
	}
}
