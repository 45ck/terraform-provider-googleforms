// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

// Package resourcedrivefile implements the googleforms_drive_file Terraform resource.
package resourcedrivefile

import "github.com/hashicorp/terraform-plugin-framework/types"

// DriveFileResourceModel describes the Terraform state for googleforms_drive_file.
type DriveFileResourceModel struct {
	ID types.String `tfsdk:"id"`

	FileID            types.String `tfsdk:"file_id"`
	Name              types.String `tfsdk:"name"`
	FolderID          types.String `tfsdk:"folder_id"`
	DeleteOnDestroy   types.Bool   `tfsdk:"delete_on_destroy"`
	SupportsAllDrives types.Bool   `tfsdk:"supports_all_drives"`

	ParentIDs types.List   `tfsdk:"parent_ids"`
	URL       types.String `tfsdk:"url"`
	MimeType  types.String `tfsdk:"mime_type"`
	Trashed   types.Bool   `tfsdk:"trashed"`
}
