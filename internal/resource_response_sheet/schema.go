// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceresponsesheet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Schema defines the Terraform schema for google_forms_response_sheet.
func (r *ResponseSheetResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Links a Google Form to a Google Spreadsheet for response collection. " +
			"Note: The Google Forms REST API v1 does not support programmatic linking of " +
			"response destinations. This resource tracks the association in Terraform state; " +
			"the actual linking must be configured manually in the Google Forms UI or via Apps Script.",
		Attributes: responseSheetAttributes(),
	}
}

// responseSheetAttributes returns the top-level attribute definitions.
func responseSheetAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "Composite ID in the format formID#spreadsheetID.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"form_id": schema.StringAttribute{
			Required:    true,
			Description: "The Google Form ID to link.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"spreadsheet_id": schema.StringAttribute{
			Required:    true,
			Description: "The Google Spreadsheet ID to link as the response destination.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"spreadsheet_url": schema.StringAttribute{
			Computed:    true,
			Description: "The URL of the linked spreadsheet.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}
