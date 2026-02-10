// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

// Package resourcesheetvalues implements the googleforms_sheet_values Terraform resource.
package resourcesheetvalues

import "github.com/hashicorp/terraform-plugin-framework/types"

// SheetValuesRowModel represents a row of string cells.
type SheetValuesRowModel struct {
	Cells types.List `tfsdk:"cells"`
}

// SheetValuesResourceModel describes the Terraform state for googleforms_sheet_values.
type SheetValuesResourceModel struct {
	ID               types.String `tfsdk:"id"`
	SpreadsheetID    types.String `tfsdk:"spreadsheet_id"`
	Range            types.String `tfsdk:"range"`
	ValueInputOption types.String `tfsdk:"value_input_option"`
	ReadBack         types.Bool   `tfsdk:"read_back"`

	Rows types.List `tfsdk:"rows"`

	UpdatedRange types.String `tfsdk:"updated_range"`
}
