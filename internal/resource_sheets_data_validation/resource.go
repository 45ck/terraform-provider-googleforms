// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsdatavalidation

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

var (
	_ resource.Resource                = &DataValidationResource{}
	_ resource.ResourceWithImportState = &DataValidationResource{}
)

// DataValidationResource implements googleforms_sheets_data_validation.
type DataValidationResource struct {
	client *client.Client
}

func NewDataValidationResource() resource.Resource {
	return &DataValidationResource{}
}

func (r *DataValidationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sheets_data_validation"
}

func (r *DataValidationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
