// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package datasourcedrivefile

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

var _ datasource.DataSource = &DriveFileDataSource{}

type DriveFileDataSource struct {
	client *client.Client
}

func NewDriveFileDataSource() datasource.DataSource {
	return &DriveFileDataSource{}
}

func (d *DriveFileDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_drive_file"
}

func (d *DriveFileDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a Google Drive file by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Drive file ID.",
			},
			"supports_all_drives": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to support shared drives when reading metadata.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "File name.",
			},
			"mime_type": schema.StringAttribute{
				Computed:    true,
				Description: "Drive mimeType.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "Drive webViewLink URL (if available).",
			},
			"trashed": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the file is trashed.",
			},
			"parent_ids": schema.ListAttribute{
				Computed:    true,
				Description: "Current Drive parent folder IDs (best-effort).",
				ElementType: types.StringType,
			},
		},
	}
}

func (d *DriveFileDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", "Expected *client.Client, got unexpected type.")
		return
	}
	d.client = c
}

func (d *DriveFileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DriveFileDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	supportsAllDrives := false
	if !data.SupportsAllDrives.IsNull() && !data.SupportsAllDrives.IsUnknown() {
		supportsAllDrives = data.SupportsAllDrives.ValueBool()
	}
	data.SupportsAllDrives = types.BoolValue(supportsAllDrives)

	f, err := d.client.Drive.GetFile(ctx, data.ID.ValueString(), supportsAllDrives)
	if err != nil {
		resp.Diagnostics.AddError("Read Drive File Failed", err.Error())
		return
	}

	data.ID = types.StringValue(f.Id)
	data.Name = types.StringValue(f.Name)
	data.MimeType = types.StringValue(f.MimeType)
	data.Trashed = types.BoolValue(f.Trashed)
	if strings.TrimSpace(f.WebViewLink) != "" {
		data.URL = types.StringValue(f.WebViewLink)
	} else {
		data.URL = types.StringNull()
	}

	if parents, err := d.client.Drive.GetParents(ctx, f.Id, supportsAllDrives); err == nil {
		lv, diags := types.ListValueFrom(ctx, types.StringType, parents)
		resp.Diagnostics.Append(diags...)
		data.ParentIDs = lv
	} else {
		resp.Diagnostics.AddWarning("Drive Parents Unavailable", err.Error())
		data.ParentIDs = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
