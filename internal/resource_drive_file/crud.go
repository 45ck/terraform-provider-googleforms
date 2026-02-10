// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcedrivefile

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	drive "google.golang.org/api/drive/v3"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

func (r *DriveFileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DriveFileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	supportsAllDrives := false
	if !plan.SupportsAllDrives.IsNull() && !plan.SupportsAllDrives.IsUnknown() {
		supportsAllDrives = plan.SupportsAllDrives.ValueBool()
	}

	fileID := plan.FileID.ValueString()
	f, err := r.client.Drive.GetFile(ctx, fileID, supportsAllDrives)
	if err != nil {
		resp.Diagnostics.AddError("Drive File Not Found", err.Error())
		return
	}

	// Apply desired rename/move on create so state matches config.
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() && strings.TrimSpace(plan.Name.ValueString()) != "" && plan.Name.ValueString() != f.Name {
		_, err := r.client.Drive.UpdateFile(ctx, fileID, &drive.File{Name: plan.Name.ValueString()}, "", "", supportsAllDrives)
		if err != nil {
			resp.Diagnostics.AddError("Rename Drive File Failed", err.Error())
			return
		}
		f.Name = plan.Name.ValueString()
	}

	if !plan.FolderID.IsNull() && !plan.FolderID.IsUnknown() {
		desiredFolder := plan.FolderID.ValueString()
		parents, perr := r.client.Drive.GetParents(ctx, fileID, supportsAllDrives)
		if perr == nil {
			cur := ""
			if len(parents) > 0 {
				cur = parents[0]
			}
			if desiredFolder != cur {
				if err := r.client.Drive.MoveToFolder(ctx, fileID, desiredFolder, supportsAllDrives); err != nil {
					resp.Diagnostics.AddError("Move Drive File Failed", err.Error())
					return
				}
				parents, _ = r.client.Drive.GetParents(ctx, fileID, supportsAllDrives)
				plan.ParentIDs, _ = types.ListValueFrom(ctx, types.StringType, parents)
			}
		}
	}

	plan.ID = types.StringValue(fileID)
	if strings.TrimSpace(f.WebViewLink) != "" {
		plan.URL = types.StringValue(f.WebViewLink)
	} else {
		plan.URL = types.StringNull()
	}
	plan.MimeType = types.StringValue(f.MimeType)
	plan.Trashed = types.BoolValue(f.Trashed)
	if plan.Name.IsNull() || plan.Name.IsUnknown() || strings.TrimSpace(plan.Name.ValueString()) == "" {
		plan.Name = types.StringValue(f.Name)
	}

	if parents, err := r.client.Drive.GetParents(ctx, fileID, supportsAllDrives); err == nil {
		lv, diags := types.ListValueFrom(ctx, types.StringType, parents)
		resp.Diagnostics.Append(diags...)
		plan.ParentIDs = lv
	} else {
		resp.Diagnostics.AddWarning("Drive Parents Unavailable", err.Error())
		plan.ParentIDs = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DriveFileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DriveFileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	supportsAllDrives := false
	if !state.SupportsAllDrives.IsNull() && !state.SupportsAllDrives.IsUnknown() {
		supportsAllDrives = state.SupportsAllDrives.ValueBool()
	}

	fileID := state.FileID.ValueString()
	f, err := r.client.Drive.GetFile(ctx, fileID, supportsAllDrives)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Drive File Failed", err.Error())
		return
	}

	state.ID = types.StringValue(fileID)
	state.MimeType = types.StringValue(f.MimeType)
	state.Trashed = types.BoolValue(f.Trashed)
	if strings.TrimSpace(f.WebViewLink) != "" {
		state.URL = types.StringValue(f.WebViewLink)
	} else {
		state.URL = types.StringNull()
	}

	// Only set name from API if it was not explicitly configured.
	if state.Name.IsNull() || state.Name.IsUnknown() || strings.TrimSpace(state.Name.ValueString()) == "" {
		state.Name = types.StringValue(f.Name)
	}

	if parents, err := r.client.Drive.GetParents(ctx, fileID, supportsAllDrives); err == nil {
		lv, diags := types.ListValueFrom(ctx, types.StringType, parents)
		resp.Diagnostics.Append(diags...)
		state.ParentIDs = lv
	} else {
		resp.Diagnostics.AddWarning("Drive Parents Unavailable", err.Error())
		state.ParentIDs = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DriveFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DriveFileResourceModel
	var state DriveFileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	supportsAllDrives := false
	if !plan.SupportsAllDrives.IsNull() && !plan.SupportsAllDrives.IsUnknown() {
		supportsAllDrives = plan.SupportsAllDrives.ValueBool()
	}

	fileID := state.FileID.ValueString()

	// Rename if name is explicitly set and changed.
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() && strings.TrimSpace(plan.Name.ValueString()) != "" && plan.Name.ValueString() != state.Name.ValueString() {
		_, err := r.client.Drive.UpdateFile(ctx, fileID, &drive.File{Name: plan.Name.ValueString()}, "", "", supportsAllDrives)
		if err != nil {
			resp.Diagnostics.AddError("Rename Drive File Failed", err.Error())
			return
		}
	}

	// Move if folder_id is set and changed. Empty string means root.
	if !plan.FolderID.IsNull() && !plan.FolderID.IsUnknown() && plan.FolderID.ValueString() != state.FolderID.ValueString() {
		if err := r.client.Drive.MoveToFolder(ctx, fileID, plan.FolderID.ValueString(), supportsAllDrives); err != nil {
			resp.Diagnostics.AddError("Move Drive File Failed", err.Error())
			return
		}
	}

	plan.ID = types.StringValue(fileID)

	// Refresh computed fields.
	f, err := r.client.Drive.GetFile(ctx, fileID, supportsAllDrives)
	if err == nil && f != nil {
		plan.MimeType = types.StringValue(f.MimeType)
		plan.Trashed = types.BoolValue(f.Trashed)
		if strings.TrimSpace(f.WebViewLink) != "" {
			plan.URL = types.StringValue(f.WebViewLink)
		} else {
			plan.URL = types.StringNull()
		}
	}
	if parents, err := r.client.Drive.GetParents(ctx, fileID, supportsAllDrives); err == nil {
		lv, diags := types.ListValueFrom(ctx, types.StringType, parents)
		resp.Diagnostics.Append(diags...)
		plan.ParentIDs = lv
	} else {
		resp.Diagnostics.AddWarning("Drive Parents Unavailable", err.Error())
		plan.ParentIDs = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DriveFileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DriveFileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteOnDestroy := false
	if !state.DeleteOnDestroy.IsNull() && !state.DeleteOnDestroy.IsUnknown() {
		deleteOnDestroy = state.DeleteOnDestroy.ValueBool()
	}
	if !deleteOnDestroy {
		// No-op: state-only resource by default.
		return
	}

	supportsAllDrives := false
	if !state.SupportsAllDrives.IsNull() && !state.SupportsAllDrives.IsUnknown() {
		supportsAllDrives = state.SupportsAllDrives.ValueBool()
	}
	_ = supportsAllDrives // reserved for future delete variants

	err := r.client.Drive.Delete(ctx, state.FileID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete Drive File Failed", err.Error())
		return
	}
}
