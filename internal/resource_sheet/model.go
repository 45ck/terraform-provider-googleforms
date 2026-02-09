// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

// Package resourcesheet implements the google_forms_sheet Terraform resource.
package resourcesheet

import "github.com/hashicorp/terraform-plugin-framework/types"

// SheetResourceModel describes the Terraform state for google_forms_sheet.
type SheetResourceModel struct {
	ID            types.String `tfsdk:"id"`
	SpreadsheetID types.String `tfsdk:"spreadsheet_id"`
	Title         types.String `tfsdk:"title"`
	RowCount      types.Int64  `tfsdk:"row_count"`
	ColumnCount   types.Int64  `tfsdk:"column_count"`
	SheetID       types.Int64  `tfsdk:"sheet_id"`
	Index         types.Int64  `tfsdk:"index"`
}
