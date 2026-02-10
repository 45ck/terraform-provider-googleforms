// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsbatchupdate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

// Compile-time interface checks.
var (
	_ resource.Resource = &SheetsBatchUpdateResource{}
)

// SheetsBatchUpdateResource implements the googleforms_sheets_batch_update Terraform resource.
type SheetsBatchUpdateResource struct {
	client *client.Client
}

// NewSheetsBatchUpdateResource returns a new resource factory function.
func NewSheetsBatchUpdateResource() resource.Resource {
	return &SheetsBatchUpdateResource{}
}

func (r *SheetsBatchUpdateResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_sheets_batch_update"
}

func (r *SheetsBatchUpdateResource) Configure(
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

