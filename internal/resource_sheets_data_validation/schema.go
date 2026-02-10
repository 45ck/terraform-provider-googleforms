// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsdatavalidation

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *DataValidationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Sheets data validation for a range via spreadsheets.batchUpdate.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Deterministic ID derived from spreadsheet_id and range.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"spreadsheet_id": schema.StringAttribute{
				Required:    true,
				Description: "Target spreadsheet ID.",
			},
			"range": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Grid range (0-based indices). End indices are exclusive.",
				Attributes: map[string]schema.Attribute{
					"sheet_id":           schema.Int64Attribute{Required: true, Description: "Sheet/tab ID (sheetId)."},
					"start_row_index":    schema.Int64Attribute{Required: true, Description: "Start row index (0-based)."},
					"end_row_index":      schema.Int64Attribute{Required: true, Description: "End row index (0-based, exclusive)."},
					"start_column_index": schema.Int64Attribute{Required: true, Description: "Start column index (0-based)."},
					"end_column_index":   schema.Int64Attribute{Required: true, Description: "End column index (0-based, exclusive)."},
				},
			},
			"rule_json": schema.StringAttribute{
				Required:    true,
				Description: "JSON encoding of a Sheets DataValidationRule object.",
			},
		},
	}
}
