// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcedrivefolder

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &DriveFolderResource{}
	_ resource.ResourceWithImportState = &DriveFolderResource{}
)

// DriveFolderResource implements the googleforms_drive_folder Terraform resource.
type DriveFolderResource struct {
	client *client.Client
}

func NewDriveFolderResource() resource.Resource {
	return &DriveFolderResource{}
}

func (r *DriveFolderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_drive_folder"
}

func (r *DriveFolderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
