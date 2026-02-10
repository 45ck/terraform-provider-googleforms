// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsnamedrange

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

var (
	_ resource.Resource                = &NamedRangeResource{}
	_ resource.ResourceWithImportState = &NamedRangeResource{}
)

// NamedRangeResource implements googleforms_sheets_named_range.
type NamedRangeResource struct {
	client *client.Client
}

func NewNamedRangeResource() resource.Resource {
	return &NamedRangeResource{}
}

func (r *NamedRangeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sheets_named_range"
}

func (r *NamedRangeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *client.Client, got unexpected type.")
		return
	}
	r.client = c
}
