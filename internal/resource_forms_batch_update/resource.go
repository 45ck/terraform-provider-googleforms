// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceformsbatchupdate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

// Compile-time interface checks.
var _ resource.Resource = &FormsBatchUpdateResource{}

// FormsBatchUpdateResource implements the googleforms_forms_batch_update Terraform resource.
type FormsBatchUpdateResource struct {
	client *client.Client
}

func NewFormsBatchUpdateResource() resource.Resource {
	return &FormsBatchUpdateResource{}
}

func (r *FormsBatchUpdateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_forms_batch_update"
}

func (r *FormsBatchUpdateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
