// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package datasourceform

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	forms "google.golang.org/api/forms/v1"

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

func TestFormDataSource_Metadata(t *testing.T) {
	t.Parallel()

	ds := NewFormDataSource()
	resp := &datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "googleforms"}, resp)

	if resp.TypeName != "googleforms_form" {
		t.Fatalf("unexpected type name: %q", resp.TypeName)
	}
}

func TestFormDataSource_Configure_WrongType(t *testing.T) {
	t.Parallel()

	dsIface := NewFormDataSource()
	ds, ok := dsIface.(*FormDataSource)
	if !ok {
		t.Fatalf("expected *FormDataSource, got %T", dsIface)
	}

	resp := &datasource.ConfigureResponse{}
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "not-a-client"}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostics")
	}
}

func TestFormDataSource_Read_SetsState(t *testing.T) {
	t.Parallel()

	mockForms := &testutil.MockFormsAPI{
		GetFunc: func(_ context.Context, formID string) (*forms.Form, error) {
			return &forms.Form{
				FormId: formID,
				Info: &forms.Info{
					Title:         "T",
					Description:   "D",
					DocumentTitle: "Doc T",
				},
				ResponderUri:  "https://docs.google.com/forms/d/e/mock/viewform",
				RevisionId:    "rev-1",
				LinkedSheetId: "sheet-1",
				Settings: &forms.FormSettings{
					EmailCollectionType: "VERIFIED",
					QuizSettings: &forms.QuizSettings{
						IsQuiz: true,
					},
				},
			}, nil
		},
	}

	dsIface := NewFormDataSource()
	ds, ok := dsIface.(*FormDataSource)
	if !ok {
		t.Fatalf("expected *FormDataSource, got %T", dsIface)
	}
	ds.client = &client.Client{Forms: mockForms}

	schemaResp := testSchema(t, ds)
	cfg := buildConfig(t, schemaResp, map[string]tftypes.Value{
		"id": tftypes.NewValue(tftypes.String, "fid"),
	})

	state := tfsdk.State{Schema: schemaResp.Schema}
	readResp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), datasource.ReadRequest{Config: cfg}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %s", readResp.Diagnostics)
	}
}
