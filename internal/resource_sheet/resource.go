// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &SheetResource{}
	_ resource.ResourceWithImportState = &SheetResource{}
)

// SheetResource implements the googleforms_sheet Terraform resource.
type SheetResource struct {
	client *client.Client
}

// NewSheetResource returns a new resource factory function.
func NewSheetResource() resource.Resource {
	return &SheetResource{}
}

// Metadata sets the resource type name.
func (r *SheetResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_sheet"
}

// Configure extracts the provider-configured client.
func (r *SheetResource) Configure(
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

