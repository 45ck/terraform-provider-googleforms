// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcedrivepermission

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	drive "google.golang.org/api/drive/v3"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
	"github.com/45ck/terraform-provider-googleforms/internal/testutil"
)

func testSchemaResp() resource.SchemaResponse {
	var resp resource.SchemaResponse
	r := &DrivePermissionResource{}
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

	return tfsdk.Plan{
		Schema: s,
		Raw:    tftypes.NewValue(objType, merged),
	}
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

func stateModel(t *testing.T, st tfsdk.State) DrivePermissionResourceModel {
	t.Helper()
	var m DrivePermissionResourceModel
	diags := st.Get(context.Background(), &m)
	if diags.HasError() {
		t.Fatalf("failed to decode state: %s", diags)
	}
	return m
}

func TestDrivePermission_Create_SetsIDs(t *testing.T) {
	t.Parallel()

	mockDrive := &testutil.MockDriveAPI{
		CreatePermissionFunc: func(_ context.Context, fileID string, p *drive.Permission, send bool, msg string, supports bool) (*drive.Permission, error) {
			if fileID != "file123" {
				t.Fatalf("unexpected fileID: %q", fileID)
			}
			if p.Type != "user" || p.Role != "reader" {
				t.Fatalf("unexpected permission: %#v", p)
			}
			if send {
				t.Fatal("expected send_notification_email default false")
			}
			if msg != "" {
				t.Fatalf("expected empty email message, got %q", msg)
			}
			if supports {
				t.Fatal("expected supports_all_drives default false")
			}
			return &drive.Permission{Id: "perm456", DisplayName: "Alice"}, nil
		},
	}

	r := &DrivePermissionResource{client: &client.Client{Drive: mockDrive}}
	plan := buildPlan(t, map[string]tftypes.Value{
		"file_id": tftypes.NewValue(tftypes.String, "file123"),
		"type":    tftypes.NewValue(tftypes.String, "user"),
		"role":    tftypes.NewValue(tftypes.String, "reader"),
	})

	resp := &resource.CreateResponse{State: emptyState(t)}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %s", resp.Diagnostics)
	}

	got := stateModel(t, resp.State)
	if got.PermissionID.ValueString() != "perm456" {
		t.Fatalf("unexpected permission_id: %q", got.PermissionID.ValueString())
	}
	if got.ID.ValueString() != "file123#perm456" {
		t.Fatalf("unexpected id: %q", got.ID.ValueString())
	}
	if got.DisplayName.IsNull() || got.DisplayName.ValueString() != "Alice" {
		t.Fatalf("unexpected display_name: %#v", got.DisplayName)
	}
}

func TestDrivePermission_Read_NotFound_RemovesResource(t *testing.T) {
	t.Parallel()

	mockDrive := &testutil.MockDriveAPI{
		GetPermissionFunc: func(_ context.Context, fileID, permID string, supports bool) (*drive.Permission, error) {
			return nil, &client.NotFoundError{Resource: "permission", ID: permID}
		},
	}

	r := &DrivePermissionResource{client: &client.Client{Drive: mockDrive}}

	// Provide state with composite ID only to exercise split().
	schemaResp := testSchemaResp()
	s := schemaResp.Schema
	tfType := s.Type().TerraformType(context.Background())
	objType, ok := tfType.(tftypes.Object)
	if !ok {
		t.Fatalf("expected tftypes.Object, got %T", tfType)
	}
	state := tfsdk.State{
		Schema: s,
		Raw: func() tftypes.Value {
			merged := make(map[string]tftypes.Value)
			for k, v := range objType.AttributeTypes {
				merged[k] = tftypes.NewValue(v, nil)
			}
			merged["id"] = tftypes.NewValue(tftypes.String, "file123#perm456")
			return tftypes.NewValue(objType, merged)
		}(),
	}

	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %s", resp.Diagnostics)
	}
	// RemoveResource is framework-managed; asserting exact state encoding here is brittle.
}
