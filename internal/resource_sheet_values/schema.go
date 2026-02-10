// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetvalues

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the Terraform schema for googleforms_sheet_values.
func (r *SheetValuesResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages a bounded range of values in a Google Spreadsheet using A1 notation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite ID in the format spreadsheetID#A1_RANGE.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"spreadsheet_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Google Spreadsheet to update.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"range": schema.StringAttribute{
				Required:    true,
				Description: "A1 notation range to write, e.g. `Config!A1:D20`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value_input_option": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("RAW"),
				Description: "How input data should be interpreted. Valid values include `RAW` and `USER_ENTERED`.",
			},
			"read_back": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether to read values back from the API during Read to detect drift.",
			},
			"rows": schema.ListNestedAttribute{
				Required:    true,
				Description: "Rows to write. Each row is an object containing `cells` (list of strings).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cells": schema.ListAttribute{
							Required:    true,
							ElementType: types.StringType,
							Description: "The cells in this row.",
						},
					},
				},
			},
			"updated_range": schema.StringAttribute{
				Computed:    true,
				Description: "The range that was updated, as reported by the Sheets API.",
			},
		},
	}
}
