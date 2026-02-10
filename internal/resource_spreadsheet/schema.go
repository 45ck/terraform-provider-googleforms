// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcespreadsheet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Schema defines the Terraform schema for googleforms_spreadsheet.
func (r *SpreadsheetResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages a Google Sheets spreadsheet.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The spreadsheet ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Required:    true,
				Description: "The spreadsheet title.",
			},
			"locale": schema.StringAttribute{
				Optional:    true,
				Description: "The locale of the spreadsheet (e.g. en_AU).",
			},
			"time_zone": schema.StringAttribute{
				Optional:    true,
				Description: "The time zone of the spreadsheet (e.g. Australia/Sydney).",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL to the spreadsheet.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

