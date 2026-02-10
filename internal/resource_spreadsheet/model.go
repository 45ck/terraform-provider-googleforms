// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

// Package resourcespreadsheet implements the googleforms_spreadsheet Terraform resource.
package resourcespreadsheet

import "github.com/hashicorp/terraform-plugin-framework/types"

// SpreadsheetResourceModel describes the Terraform state for googleforms_spreadsheet.
type SpreadsheetResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Title    types.String `tfsdk:"title"`
	Locale   types.String `tfsdk:"locale"`
	TimeZone types.String `tfsdk:"time_zone"`
	URL      types.String `tfsdk:"url"`
}

