// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsconditionalformatrule

import "github.com/hashicorp/terraform-plugin-framework/types"

type ConditionalFormatRuleResourceModel struct {
	ID            types.String `tfsdk:"id"`
	SpreadsheetID types.String `tfsdk:"spreadsheet_id"`
	SheetID       types.Int64  `tfsdk:"sheet_id"`
	Index         types.Int64  `tfsdk:"index"`
	RuleJSON      types.String `tfsdk:"rule_json"`
}
