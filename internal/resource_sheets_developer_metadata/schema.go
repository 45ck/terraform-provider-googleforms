// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsdevelopermetadata

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *DeveloperMetadataResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Google Sheets developer metadata via spreadsheets.batchUpdate.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite ID in the format spreadsheetID#metadataId.",
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
			"metadata_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Developer metadata ID assigned by the API.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"metadata_key": schema.StringAttribute{
				Required:    true,
				Description: "Developer metadata key.",
			},
			"metadata_value": schema.StringAttribute{
				Optional:    true,
				Description: "Developer metadata value.",
			},
			"visibility": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("DOCUMENT"),
				Description: "Visibility: DOCUMENT (default) or PROJECT.",
				Validators: []validator.String{
					stringvalidator.OneOf("DOCUMENT", "PROJECT"),
				},
			},
			"location": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Optional sheet-scoped location. If omitted, metadata is spreadsheet-scoped.",
				Attributes: map[string]schema.Attribute{
					"sheet_id": schema.Int64Attribute{
						Optional:    true,
						Description: "Sheet/tab ID (sheetId) to scope the metadata to the sheet.",
					},
				},
			},
		},
	}
}
