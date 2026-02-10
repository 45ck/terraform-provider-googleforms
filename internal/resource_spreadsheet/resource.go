// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcespreadsheet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &SpreadsheetResource{}
	_ resource.ResourceWithImportState = &SpreadsheetResource{}
)

// SpreadsheetResource implements the googleforms_spreadsheet Terraform resource.
type SpreadsheetResource struct {
	client *client.Client
}

// NewSpreadsheetResource returns a new resource factory function.
func NewSpreadsheetResource() resource.Resource {
	return &SpreadsheetResource{}
}

// Metadata sets the resource type name.
func (r *SpreadsheetResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_spreadsheet"
}

// Configure extracts the provider-configured client.
func (r *SpreadsheetResource) Configure(
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

