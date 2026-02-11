// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package datasourceform

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

var _ datasource.DataSource = &FormDataSource{}

type FormDataSource struct {
	client *client.Client
}

func NewFormDataSource() datasource.DataSource {
	return &FormDataSource{}
}

func (d *FormDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_form"
}

func (d *FormDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a Google Form by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Form ID.",
			},
			"title": schema.StringAttribute{
				Computed:    true,
				Description: "Form title.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Form description.",
			},
			"responder_uri": schema.StringAttribute{
				Computed:    true,
				Description: "Responder URI.",
			},
			"edit_uri": schema.StringAttribute{
				Computed:    true,
				Description: "Edit URI.",
			},
			"document_title": schema.StringAttribute{
				Computed:    true,
				Description: "Drive document title.",
			},
			"revision_id": schema.StringAttribute{
				Computed:    true,
				Description: "Form revision ID.",
			},
			"linked_sheet_id": schema.StringAttribute{
				Computed:    true,
				Description: "Linked Sheet ID (if responses are linked).",
			},
			"quiz": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether quiz mode is enabled.",
			},
			"email_collection_type": schema.StringAttribute{
				Computed:    true,
				Description: "Email collection type (DO_NOT_COLLECT, VERIFIED, RESPONDER_INPUT).",
			},
		},
	}
}

func (d *FormDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FormDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FormDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	f, err := d.client.Forms.Get(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Form Failed", err.Error())
		return
	}

	data.ID = types.StringValue(f.FormId)
	if f.Info != nil {
		data.Title = types.StringValue(f.Info.Title)
		data.Description = types.StringValue(f.Info.Description)
		data.DocumentTitle = types.StringValue(f.Info.DocumentTitle)
	}

	data.ResponderURI = types.StringValue(f.ResponderUri)
	if f.FormId != "" {
		data.EditURI = types.StringValue("https://docs.google.com/forms/d/" + f.FormId + "/edit")
	}
	data.RevisionID = types.StringValue(f.RevisionId)
	data.LinkedSheetID = types.StringValue(f.LinkedSheetId)

	if f.Settings != nil && f.Settings.QuizSettings != nil {
		data.Quiz = types.BoolValue(f.Settings.QuizSettings.IsQuiz)
	} else {
		data.Quiz = types.BoolValue(false)
	}
	if f.Settings != nil && f.Settings.EmailCollectionType != "" {
		data.EmailCollectionType = types.StringValue(f.Settings.EmailCollectionType)
	} else {
		data.EmailCollectionType = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
