// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsbatchupdate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Schema defines the Terraform schema for googleforms_sheets_batch_update.
func (r *SheetsBatchUpdateResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Escape hatch: applies raw Google Sheets `spreadsheets.batchUpdate` requests to a spreadsheet.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Deterministic ID derived from the spreadsheet ID and request JSON.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"spreadsheet_id": schema.StringAttribute{
				Required:    true,
				Description: "The target spreadsheet ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"requests_json": schema.StringAttribute{
				Required: true,
				Description: "JSON defining either an array of Sheets `Request` objects, or an object matching " +
					"`BatchUpdateSpreadsheetRequest` (e.g. `{ \"requests\": [...] }`).",
			},
			"include_spreadsheet_in_response": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "If true, asks the API to include the updated spreadsheet in the response (can be large).",
			},
			"store_response_json": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "If true, stores the API response JSON in state (can be large).",
			},
			"response_json": schema.StringAttribute{
				Computed:    true,
				Description: "JSON-encoded `BatchUpdateSpreadsheetResponse` (only set when store_response_json is true).",
			},
		},
	}
}

