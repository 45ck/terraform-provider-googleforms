// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsnamedrange

import "github.com/hashicorp/terraform-plugin-framework/types"

type GridRangeModel struct {
	SheetID          types.Int64 `tfsdk:"sheet_id"`
	StartRowIndex    types.Int64 `tfsdk:"start_row_index"`
	EndRowIndex      types.Int64 `tfsdk:"end_row_index"`
	StartColumnIndex types.Int64 `tfsdk:"start_column_index"`
	EndColumnIndex   types.Int64 `tfsdk:"end_column_index"`
}

type NamedRangeResourceModel struct {
	ID            types.String   `tfsdk:"id"`
	SpreadsheetID types.String   `tfsdk:"spreadsheet_id"`
	NamedRangeID  types.String   `tfsdk:"named_range_id"`
	Name          types.String   `tfsdk:"name"`
	Range         GridRangeModel `tfsdk:"range"`
}
