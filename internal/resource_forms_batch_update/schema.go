// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceformsbatchupdate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Schema defines the Terraform schema for googleforms_forms_batch_update.
func (r *FormsBatchUpdateResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Escape hatch: applies raw Google Forms `forms.batchUpdate` requests to a form.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Deterministic ID derived from the form ID and request JSON.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"form_id": schema.StringAttribute{
				Required:    true,
				Description: "The target form ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"requests_json": schema.StringAttribute{
				Required: true,
				Description: "JSON defining either an array of Forms `Request` objects, or an object matching " +
					"`BatchUpdateFormRequest` (e.g. `{ \"requests\": [...] }`).",
			},
			"include_form_in_response": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "If true, asks the API to include the updated form in the response.",
			},
			"required_revision_id": schema.StringAttribute{
				Optional:    true,
				Description: "If set, uses write control (requiredRevisionId) and errors if the form revision has changed since this ID.",
			},
			"store_response_json": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "If true, stores the API response JSON in state (can be large).",
			},
			"response_json": schema.StringAttribute{
				Computed:    true,
				Description: "JSON-encoded `BatchUpdateFormResponse` (only set when store_response_json is true).",
			},
		},
	}
}
