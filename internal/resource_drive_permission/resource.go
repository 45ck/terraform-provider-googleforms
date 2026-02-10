// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcedrivepermission

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &DrivePermissionResource{}
	_ resource.ResourceWithImportState = &DrivePermissionResource{}
)

// DrivePermissionResource implements the googleforms_drive_permission Terraform resource.
type DrivePermissionResource struct {
	client *client.Client
}

// NewDrivePermissionResource returns a new resource factory function.
func NewDrivePermissionResource() resource.Resource {
	return &DrivePermissionResource{}
}

func (r *DrivePermissionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_drive_permission"
}

func (r *DrivePermissionResource) Configure(
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
