// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsdevelopermetadata

import "github.com/hashicorp/terraform-plugin-framework/types"

type MetadataLocationModel struct {
	SheetID types.Int64 `tfsdk:"sheet_id"`
}

type DeveloperMetadataResourceModel struct {
	ID            types.String           `tfsdk:"id"`
	SpreadsheetID types.String           `tfsdk:"spreadsheet_id"`
	MetadataID    types.Int64            `tfsdk:"metadata_id"`
	MetadataKey   types.String           `tfsdk:"metadata_key"`
	MetadataValue types.String           `tfsdk:"metadata_value"`
	Visibility    types.String           `tfsdk:"visibility"`
	Location      *MetadataLocationModel `tfsdk:"location"`
}
