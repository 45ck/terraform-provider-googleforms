// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceresponsesheet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

// Compile-time interface checks.
var _ resource.Resource = &ResponseSheetResource{}

// ResponseSheetResource implements the google_forms_response_sheet Terraform resource.
type ResponseSheetResource struct {
	client *client.Client
}

// NewResponseSheetResource returns a new resource factory function.
func NewResponseSheetResource() resource.Resource {
	return &ResponseSheetResource{}
}

// Metadata sets the resource type name.
func (r *ResponseSheetResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_response_sheet"
}

// Configure extracts the provider-configured client.
func (r *ResponseSheetResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *client.Client, got unexpected type.",
		)
		return
	}

	r.client = c
}
