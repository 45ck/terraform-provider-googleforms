// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

// Package resourceresponsesheet implements the googleforms_response_sheet
// Terraform resource.
package resourceresponsesheet

import "github.com/hashicorp/terraform-plugin-framework/types"

// ResponseSheetResourceModel describes the Terraform state for
// googleforms_response_sheet.
type ResponseSheetResourceModel struct {
	ID             types.String `tfsdk:"id"`
	FormID         types.String `tfsdk:"form_id"`
	SpreadsheetID  types.String `tfsdk:"spreadsheet_id"`
	SpreadsheetURL types.String `tfsdk:"spreadsheet_url"`
}

