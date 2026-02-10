// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package datasourcespreadsheet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

var (
	_ datasource.DataSource = &SpreadsheetDataSource{}
)

// SpreadsheetDataSource implements the googleforms_spreadsheet data source.
type SpreadsheetDataSource struct {
	client *client.Client
}

func NewSpreadsheetDataSource() datasource.DataSource {
	return &SpreadsheetDataSource{}
}

func (d *SpreadsheetDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_spreadsheet"
}

func (d *SpreadsheetDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a Google Sheets spreadsheet by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Spreadsheet ID.",
			},
			"title": schema.StringAttribute{
				Computed:    true,
				Description: "Spreadsheet title.",
			},
			"locale": schema.StringAttribute{
				Computed:    true,
				Description: "Spreadsheet locale.",
			},
			"time_zone": schema.StringAttribute{
				Computed:    true,
				Description: "Spreadsheet time zone.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "Spreadsheet URL.",
			},
		},
	}
}

func (d *SpreadsheetDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected *client.Client, got unexpected type.",
		)
		return
	}

	d.client = c
}

func (d *SpreadsheetDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data SpreadsheetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ss, err := d.client.Sheets.Get(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Spreadsheet Failed", err.Error())
		return
	}

	data.ID = types.StringValue(ss.SpreadsheetId)
	data.URL = types.StringValue(ss.SpreadsheetUrl)

	if ss.Properties != nil {
		data.Title = types.StringValue(ss.Properties.Title)
		if ss.Properties.Locale != "" {
			data.Locale = types.StringValue(ss.Properties.Locale)
		}
		if ss.Properties.TimeZone != "" {
			data.TimeZone = types.StringValue(ss.Properties.TimeZone)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

