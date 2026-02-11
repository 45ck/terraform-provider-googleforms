// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package datasourcedrivefile

import "github.com/hashicorp/terraform-plugin-framework/types"

type DriveFileDataSourceModel struct {
	ID types.String `tfsdk:"id"`

	SupportsAllDrives types.Bool `tfsdk:"supports_all_drives"`

	Name      types.String `tfsdk:"name"`
	MimeType  types.String `tfsdk:"mime_type"`
	URL       types.String `tfsdk:"url"`
	Trashed   types.Bool   `tfsdk:"trashed"`
	ParentIDs types.List   `tfsdk:"parent_ids"`
}
