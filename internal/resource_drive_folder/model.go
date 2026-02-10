// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

// Package resourcedrivefolder implements the googleforms_drive_folder Terraform resource.
package resourcedrivefolder

import "github.com/hashicorp/terraform-plugin-framework/types"

// DriveFolderResourceModel describes the Terraform state for googleforms_drive_folder.
type DriveFolderResourceModel struct {
	ID types.String `tfsdk:"id"`

	Name              types.String `tfsdk:"name"`
	ParentID          types.String `tfsdk:"parent_id"`
	SupportsAllDrives types.Bool   `tfsdk:"supports_all_drives"`

	ParentIDs types.List   `tfsdk:"parent_ids"`
	URL       types.String `tfsdk:"url"`
}
