// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package datasourcedrivefile

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	drive "google.golang.org/api/drive/v3"

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

func TestDriveFileDataSource_Metadata(t *testing.T) {
	t.Parallel()

	ds := NewDriveFileDataSource()
	resp := &datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "googleforms"}, resp)

	if resp.TypeName != "googleforms_drive_file" {
		t.Fatalf("unexpected type name: %q", resp.TypeName)
	}
}

func TestDriveFileDataSource_Configure_WrongType(t *testing.T) {
	t.Parallel()

	dsIface := NewDriveFileDataSource()
	ds, ok := dsIface.(*DriveFileDataSource)
	if !ok {
		t.Fatalf("expected *DriveFileDataSource, got %T", dsIface)
	}

	resp := &datasource.ConfigureResponse{}
	ds.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "not-a-client"}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostics")
	}
}

func TestDriveFileDataSource_Read_SetsState(t *testing.T) {
	t.Parallel()

	mockDrive := &testutil.MockDriveAPI{
		GetFileFunc: func(_ context.Context, fileID string, supportsAllDrives bool) (*drive.File, error) {
			if supportsAllDrives {
				t.Fatalf("expected supportsAllDrives=false by default")
			}
			return &drive.File{
				Id:          fileID,
				Name:        "N",
				MimeType:    "application/vnd.google-apps.document",
				WebViewLink: "https://drive.google.com/file/d/" + fileID + "/view",
				Trashed:     false,
			}, nil
		},
		GetParentsFunc: func(_ context.Context, fileID string, supportsAllDrives bool) ([]string, error) {
			if supportsAllDrives {
				t.Fatalf("expected supportsAllDrives=false by default")
			}
			if fileID != "did" {
				t.Fatalf("unexpected fileID: %q", fileID)
			}
			return []string{"p1", "p2"}, nil
		},
	}

	dsIface := NewDriveFileDataSource()
	ds, ok := dsIface.(*DriveFileDataSource)
	if !ok {
		t.Fatalf("expected *DriveFileDataSource, got %T", dsIface)
	}
	ds.client = &client.Client{Drive: mockDrive}

	schemaResp := testSchema(t, ds)
	cfg := buildConfig(t, schemaResp, map[string]tftypes.Value{
		"id": tftypes.NewValue(tftypes.String, "did"),
	})

	state := tfsdk.State{Schema: schemaResp.Schema}
	readResp := &datasource.ReadResponse{State: state}
	ds.Read(context.Background(), datasource.ReadRequest{Config: cfg}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %s", readResp.Diagnostics)
	}
}
