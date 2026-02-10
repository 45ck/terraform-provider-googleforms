// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcedrivefile

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &DriveFileResource{}
	_ resource.ResourceWithImportState = &DriveFileResource{}
)

// DriveFileResource manages selected metadata of an existing Drive file.
//
// This resource does not create Drive files. It adopts an existing file by ID
// and optionally renames/moves it.
type DriveFileResource struct {
	client *client.Client
}

func NewDriveFileResource() resource.Resource {
	return &DriveFileResource{}
}

func (r *DriveFileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_drive_file"
}

func (r *DriveFileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
