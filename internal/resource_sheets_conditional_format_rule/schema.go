// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsconditionalformatrule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *ConditionalFormatRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Google Sheets conditional format rule. Note: Rules are addressed by index and can be affected by out-of-band edits.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite ID in the format spreadsheetID#sheetId#index.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"spreadsheet_id": schema.StringAttribute{
				Required:    true,
				Description: "Target spreadsheet ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sheet_id": schema.Int64Attribute{
				Required:    true,
				Description: "Sheet/tab ID (sheetId).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"index": schema.Int64Attribute{
				Required:    true,
				Description: "0-based rule index within the sheet's conditionalFormats array.",
			},
			"rule_json": schema.StringAttribute{
				Required:    true,
				Description: "JSON encoding of a Sheets ConditionalFormatRule object.",
			},
		},
	}
}
