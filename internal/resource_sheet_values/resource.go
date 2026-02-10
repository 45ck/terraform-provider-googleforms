// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetvalues

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &SheetValuesResource{}
	_ resource.ResourceWithImportState = &SheetValuesResource{}
)

// SheetValuesResource implements the google_forms_sheet_values Terraform resource.
type SheetValuesResource struct {
	client *client.Client
}

// NewSheetValuesResource returns a new resource factory function.
func NewSheetValuesResource() resource.Resource {
	return &SheetValuesResource{}
}

func (r *SheetValuesResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_sheet_values"
}

func (r *SheetValuesResource) Configure(
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
