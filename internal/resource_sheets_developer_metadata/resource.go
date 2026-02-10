// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsdevelopermetadata

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

var (
	_ resource.Resource                = &DeveloperMetadataResource{}
	_ resource.ResourceWithImportState = &DeveloperMetadataResource{}
)

// DeveloperMetadataResource implements googleforms_sheets_developer_metadata.
type DeveloperMetadataResource struct {
	client *client.Client
}

func NewDeveloperMetadataResource() resource.Resource {
	return &DeveloperMetadataResource{}
}

func (r *DeveloperMetadataResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sheets_developer_metadata"
}

func (r *DeveloperMetadataResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
