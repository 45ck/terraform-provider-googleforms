// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

// Package resourcesheetsbatchupdate implements the google_forms_sheets_batch_update Terraform resource.
package resourcesheetsbatchupdate

import "github.com/hashicorp/terraform-plugin-framework/types"

// SheetsBatchUpdateResourceModel describes the Terraform state for google_forms_sheets_batch_update.
type SheetsBatchUpdateResourceModel struct {
	ID types.String `tfsdk:"id"`

	SpreadsheetID types.String `tfsdk:"spreadsheet_id"`

	RequestsJSON                 types.String `tfsdk:"requests_json"`
	IncludeSpreadsheetInResponse types.Bool   `tfsdk:"include_spreadsheet_in_response"`
	StoreResponseJSON            types.Bool   `tfsdk:"store_response_json"`

	ResponseJSON types.String `tfsdk:"response_json"`
}
