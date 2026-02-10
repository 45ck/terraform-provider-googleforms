// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package datasourcesheetvalues

import "github.com/hashicorp/terraform-plugin-framework/types"

type SheetValuesRowModel struct {
	Cells types.List `tfsdk:"cells"`
}

// SheetValuesDataSourceModel describes the Terraform state for googleforms_sheet_values data source.
type SheetValuesDataSourceModel struct {
	SpreadsheetID types.String `tfsdk:"spreadsheet_id"`
	Range         types.String `tfsdk:"range"`
	Rows          types.List   `tfsdk:"rows"`
}

