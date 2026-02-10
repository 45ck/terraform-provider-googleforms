// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

// Package resourcedrivepermission implements the googleforms_drive_permission Terraform resource.
package resourcedrivepermission

import "github.com/hashicorp/terraform-plugin-framework/types"

// DrivePermissionResourceModel describes the Terraform state for googleforms_drive_permission.
type DrivePermissionResourceModel struct {
	ID types.String `tfsdk:"id"`

	FileID         types.String `tfsdk:"file_id"`
	PermissionID   types.String `tfsdk:"permission_id"`
	Type           types.String `tfsdk:"type"`
	Role           types.String `tfsdk:"role"`
	EmailAddress   types.String `tfsdk:"email_address"`
	Domain         types.String `tfsdk:"domain"`
	AllowDiscovery types.Bool   `tfsdk:"allow_file_discovery"`

	SendNotificationEmail types.Bool   `tfsdk:"send_notification_email"`
	EmailMessage          types.String `tfsdk:"email_message"`
	SupportsAllDrives     types.Bool   `tfsdk:"supports_all_drives"`

	DisplayName types.String `tfsdk:"display_name"`
}

