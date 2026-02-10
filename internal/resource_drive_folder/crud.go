// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcedrivefolder

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	drive "google.golang.org/api/drive/v3"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

const folderMimeType = "application/vnd.google-apps.folder"

func (r *DriveFolderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DriveFolderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	supportsAllDrives := false
	if !plan.SupportsAllDrives.IsNull() && !plan.SupportsAllDrives.IsUnknown() {
		supportsAllDrives = plan.SupportsAllDrives.ValueBool()
	}

	f := &drive.File{
		Name:     plan.Name.ValueString(),
		MimeType: folderMimeType,
	}
	if !plan.ParentID.IsNull() && !plan.ParentID.IsUnknown() && strings.TrimSpace(plan.ParentID.ValueString()) != "" {
		f.Parents = []string{plan.ParentID.ValueString()}
	}

	created, err := r.client.Drive.CreateFile(ctx, f, supportsAllDrives)
	if err != nil {
		resp.Diagnostics.AddError("Create Drive Folder Failed", err.Error())
		return
	}

	plan.ID = types.StringValue(created.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.URL = types.StringValue(strings.TrimSpace(created.WebViewLink))
	if strings.TrimSpace(created.WebViewLink) == "" {
		plan.URL = types.StringNull()
	}

	if parents, err := r.client.Drive.GetParents(ctx, created.Id, supportsAllDrives); err == nil {
		lv, diags := types.ListValueFrom(ctx, types.StringType, parents)
		resp.Diagnostics.Append(diags...)
		plan.ParentIDs = lv
	} else {
		resp.Diagnostics.AddWarning("Drive Parents Unavailable", err.Error())
		plan.ParentIDs = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	tflog.Debug(ctx, "created drive folder", map[string]interface{}{"id": plan.ID.ValueString()})
}

func (r *DriveFolderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DriveFolderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	supportsAllDrives := false
	if !state.SupportsAllDrives.IsNull() && !state.SupportsAllDrives.IsUnknown() {
		supportsAllDrives = state.SupportsAllDrives.ValueBool()
	}

	f, err := r.client.Drive.GetFile(ctx, state.ID.ValueString(), supportsAllDrives)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Drive Folder Failed", err.Error())
		return
	}
	if f != nil && strings.TrimSpace(f.MimeType) != "" && f.MimeType != folderMimeType {
		resp.Diagnostics.AddError("Not A Folder", fmt.Sprintf("Drive file %s is not a folder (mimeType=%q)", state.ID.ValueString(), f.MimeType))
		return
	}

	state.Name = types.StringValue(f.Name)
	if strings.TrimSpace(f.WebViewLink) != "" {
		state.URL = types.StringValue(f.WebViewLink)
	} else {
		state.URL = types.StringNull()
	}

	if parents, err := r.client.Drive.GetParents(ctx, state.ID.ValueString(), supportsAllDrives); err == nil {
		lv, diags := types.ListValueFrom(ctx, types.StringType, parents)
		resp.Diagnostics.Append(diags...)
		state.ParentIDs = lv
	} else {
		resp.Diagnostics.AddWarning("Drive Parents Unavailable", err.Error())
		state.ParentIDs = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DriveFolderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DriveFolderResourceModel
	var state DriveFolderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	supportsAllDrives := false
	if !plan.SupportsAllDrives.IsNull() && !plan.SupportsAllDrives.IsUnknown() {
		supportsAllDrives = plan.SupportsAllDrives.ValueBool()
	}

	// Rename if changed.
	if plan.Name.ValueString() != state.Name.ValueString() {
		_, err := r.client.Drive.UpdateFile(ctx, state.ID.ValueString(), &drive.File{Name: plan.Name.ValueString()}, "", "", supportsAllDrives)
		if err != nil {
			resp.Diagnostics.AddError("Update Drive Folder Failed", err.Error())
			return
		}
	}

	// Move if parent_id changed (best-effort). This treats parent_id as the desired sole parent.
	planParent := ""
	if !plan.ParentID.IsNull() && !plan.ParentID.IsUnknown() {
		planParent = plan.ParentID.ValueString()
	}
	stateParent := ""
	if !state.ParentID.IsNull() && !state.ParentID.IsUnknown() {
		stateParent = state.ParentID.ValueString()
	}

	if strings.TrimSpace(planParent) != "" && planParent != stateParent {
		if err := r.client.Drive.MoveToFolder(ctx, state.ID.ValueString(), planParent, supportsAllDrives); err != nil {
			resp.Diagnostics.AddError("Move Drive Folder Failed", err.Error())
			return
		}
	}

	plan.ID = state.ID
	plan.URL = state.URL

	if parents, err := r.client.Drive.GetParents(ctx, state.ID.ValueString(), supportsAllDrives); err == nil {
		lv, diags := types.ListValueFrom(ctx, types.StringType, parents)
		resp.Diagnostics.Append(diags...)
		plan.ParentIDs = lv
	} else {
		resp.Diagnostics.AddWarning("Drive Parents Unavailable", err.Error())
		plan.ParentIDs = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DriveFolderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DriveFolderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Drive.Delete(ctx, state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete Drive Folder Failed", err.Error())
		return
	}
}

func (r *DriveFolderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
