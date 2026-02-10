// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package datasourcespreadsheet

import "github.com/hashicorp/terraform-plugin-framework/types"

// SpreadsheetDataSourceModel describes the Terraform state for googleforms_spreadsheet data source.
type SpreadsheetDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Title    types.String `tfsdk:"title"`
	Locale   types.String `tfsdk:"locale"`
	TimeZone types.String `tfsdk:"time_zone"`
	URL      types.String `tfsdk:"url"`
}
