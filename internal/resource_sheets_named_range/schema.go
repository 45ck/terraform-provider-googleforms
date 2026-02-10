// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsnamedrange

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *NamedRangeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Google Sheets named range via spreadsheets.batchUpdate.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite ID in the format spreadsheetID#namedRangeId.",
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
			"named_range_id": schema.StringAttribute{
				Computed:    true,
				Description: "The numeric named range ID assigned by the API.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the named range.",
			},
			"range": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Grid range (0-based indices). End indices are exclusive.",
				Attributes: map[string]schema.Attribute{
					"sheet_id": schema.Int64Attribute{
						Required:    true,
						Description: "Sheet/tab ID (sheetId).",
					},
					"start_row_index": schema.Int64Attribute{
						Required:    true,
						Description: "Start row index (0-based).",
					},
					"end_row_index": schema.Int64Attribute{
						Required:    true,
						Description: "End row index (0-based, exclusive).",
					},
					"start_column_index": schema.Int64Attribute{
						Required:    true,
						Description: "Start column index (0-based).",
					},
					"end_column_index": schema.Int64Attribute{
						Required:    true,
						Description: "End column index (0-based, exclusive).",
					},
				},
			},
		},
	}
}
