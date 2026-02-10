// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsprotectedrange

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

var (
	_ resource.Resource                = &ProtectedRangeResource{}
	_ resource.ResourceWithImportState = &ProtectedRangeResource{}
)

// ProtectedRangeResource implements googleforms_sheets_protected_range.
type ProtectedRangeResource struct {
	client *client.Client
}

func NewProtectedRangeResource() resource.Resource {
	return &ProtectedRangeResource{}
}

func (r *ProtectedRangeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sheets_protected_range"
}

func (r *ProtectedRangeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
