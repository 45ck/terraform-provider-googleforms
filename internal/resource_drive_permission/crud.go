// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcedrivepermission

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	drive "google.golang.org/api/drive/v3"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

func (r *DrivePermissionResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan DrivePermissionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fileID := plan.FileID.ValueString()

	p := &drive.Permission{
		Type: plan.Type.ValueString(),
		Role: plan.Role.ValueString(),
	}
	if !plan.EmailAddress.IsNull() && !plan.EmailAddress.IsUnknown() {
		p.EmailAddress = plan.EmailAddress.ValueString()
	}
	if !plan.Domain.IsNull() && !plan.Domain.IsUnknown() {
		p.Domain = plan.Domain.ValueString()
	}
	if !plan.AllowDiscovery.IsNull() && !plan.AllowDiscovery.IsUnknown() {
		p.AllowFileDiscovery = plan.AllowDiscovery.ValueBool()
	}

	emailMessage := ""
	if !plan.EmailMessage.IsNull() && !plan.EmailMessage.IsUnknown() {
		emailMessage = plan.EmailMessage.ValueString()
	}

	supportsAllDrives := false
	if !plan.SupportsAllDrives.IsNull() && !plan.SupportsAllDrives.IsUnknown() {
		supportsAllDrives = plan.SupportsAllDrives.ValueBool()
	}

	sendNotification := false
	if !plan.SendNotificationEmail.IsNull() && !plan.SendNotificationEmail.IsUnknown() {
		sendNotification = plan.SendNotificationEmail.ValueBool()
	}

	created, err := r.client.Drive.CreatePermission(ctx, fileID, p, sendNotification, emailMessage, supportsAllDrives)
	if err != nil {
		resp.Diagnostics.AddError("Create Drive Permission Failed", err.Error())
		return
	}

	plan.PermissionID = types.StringValue(created.Id)
	plan.ID = types.StringValue(composeID(fileID, created.Id))
	if strings.TrimSpace(created.DisplayName) != "" {
		plan.DisplayName = types.StringValue(created.DisplayName)
	} else {
		plan.DisplayName = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	tflog.Debug(ctx, "created drive permission", map[string]interface{}{"id": plan.ID.ValueString()})
}

func (r *DrivePermissionResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state DrivePermissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fileID, permID, diags := parseID(state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	supportsAllDrives := false
	if !state.SupportsAllDrives.IsNull() && !state.SupportsAllDrives.IsUnknown() {
		supportsAllDrives = state.SupportsAllDrives.ValueBool()
	}

	p, err := r.client.Drive.GetPermission(ctx, fileID, permID, supportsAllDrives)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Drive Permission Failed", err.Error())
		return
	}

	state.FileID = types.StringValue(fileID)
	state.PermissionID = types.StringValue(permID)
	state.ID = types.StringValue(composeID(fileID, permID))
	if strings.TrimSpace(p.Type) != "" {
		state.Type = types.StringValue(p.Type)
	}
	if strings.TrimSpace(p.Role) != "" {
		state.Role = types.StringValue(p.Role)
	}
	if strings.TrimSpace(p.EmailAddress) != "" {
		state.EmailAddress = types.StringValue(p.EmailAddress)
	}
	if strings.TrimSpace(p.Domain) != "" {
		state.Domain = types.StringValue(p.Domain)
	}
	state.AllowDiscovery = types.BoolValue(p.AllowFileDiscovery)
	if strings.TrimSpace(p.DisplayName) != "" {
		state.DisplayName = types.StringValue(p.DisplayName)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DrivePermissionResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// All configurable attributes are RequiresReplace in this MVP.
	var plan DrivePermissionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DrivePermissionResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state DrivePermissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fileID, permID, diags := parseID(state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	supportsAllDrives := false
	if !state.SupportsAllDrives.IsNull() && !state.SupportsAllDrives.IsUnknown() {
		supportsAllDrives = state.SupportsAllDrives.ValueBool()
	}

	err := r.client.Drive.DeletePermission(ctx, fileID, permID, supportsAllDrives)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete Drive Permission Failed", err.Error())
		return
	}
}

func (r *DrivePermissionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Import format: fileID#permissionID
	fileID, permID, diags := split(req.ID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("file_id"), types.StringValue(fileID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("permission_id"), types.StringValue(permID))...)
}

func composeID(fileID, permID string) string {
	return fmt.Sprintf("%s#%s", fileID, permID)
}

func split(id string) (string, string, diag.Diagnostics) {
	var diags diag.Diagnostics
	parts := strings.SplitN(id, "#", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		diags.AddError("Invalid Import ID", fmt.Sprintf("Expected import ID format fileID#permissionID, got %q", id))
		return "", "", diags
	}
	return parts[0], parts[1], diags
}

func parseID(state DrivePermissionResourceModel) (string, string, diag.Diagnostics) {
	if !state.FileID.IsNull() && !state.FileID.IsUnknown() && !state.PermissionID.IsNull() && !state.PermissionID.IsUnknown() {
		return state.FileID.ValueString(), state.PermissionID.ValueString(), diag.Diagnostics{}
	}
	return split(state.ID.ValueString())
}
