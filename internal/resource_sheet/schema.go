// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Schema defines the Terraform schema for google_forms_sheet.
func (r *SheetResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages an individual sheet (tab) within a Google Spreadsheet.",
		Attributes:  sheetAttributes(),
	}
}

// sheetAttributes returns the top-level attribute definitions.
func sheetAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "Composite ID in the format spreadsheetID#sheetID.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"spreadsheet_id": schema.StringAttribute{
			Required:    true,
			Description: "The ID of the Google Spreadsheet that contains this sheet.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"title": schema.StringAttribute{
			Required:    true,
			Description: "The title (name) of the sheet tab.",
		},
		"row_count": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(1000),
			Description: "The number of rows in the sheet grid. Defaults to 1000.",
		},
		"column_count": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(26),
			Description: "The number of columns in the sheet grid. Defaults to 26.",
		},
		"sheet_id": schema.Int64Attribute{
			Computed:    true,
			Description: "Google's internal sheet ID within the spreadsheet.",
		},
		"index": schema.Int64Attribute{
			Computed:    true,
			Description: "The zero-based position of the sheet in the spreadsheet.",
		},
	}
}
