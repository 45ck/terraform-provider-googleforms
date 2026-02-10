// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsprotectedrange

import "github.com/hashicorp/terraform-plugin-framework/types"

type GridRangeModel struct {
	SheetID          types.Int64 `tfsdk:"sheet_id"`
	StartRowIndex    types.Int64 `tfsdk:"start_row_index"`
	EndRowIndex      types.Int64 `tfsdk:"end_row_index"`
	StartColumnIndex types.Int64 `tfsdk:"start_column_index"`
	EndColumnIndex   types.Int64 `tfsdk:"end_column_index"`
}

type ProtectedRangeResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	SpreadsheetID    types.String   `tfsdk:"spreadsheet_id"`
	ProtectedRangeID types.Int64    `tfsdk:"protected_range_id"`
	Range            GridRangeModel `tfsdk:"range"`
	Description      types.String   `tfsdk:"description"`
	WarningOnly      types.Bool     `tfsdk:"warning_only"`
}
