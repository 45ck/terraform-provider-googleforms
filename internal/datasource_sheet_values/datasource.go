// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package datasourcesheetvalues

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sheets "google.golang.org/api/sheets/v4"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

var _ datasource.DataSource = &SheetValuesDataSource{}

// SheetValuesDataSource implements the googleforms_sheet_values data source.
type SheetValuesDataSource struct {
	client *client.Client
}

func NewSheetValuesDataSource() datasource.DataSource {
	return &SheetValuesDataSource{}
}

func (d *SheetValuesDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_sheet_values"
}

func (d *SheetValuesDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Reads a bounded A1 range of values from a Google Spreadsheet.",
		Attributes: map[string]schema.Attribute{
			"spreadsheet_id": schema.StringAttribute{
				Required:    true,
				Description: "Spreadsheet ID.",
			},
			"range": schema.StringAttribute{
				Required:    true,
				Description: "A1 range to read, e.g. `Config!A1:D20`.",
			},
			"rows": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Rows returned from the API. Each row is an object containing `cells` (list of strings).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cells": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "The cells in this row.",
						},
					},
				},
			},
		},
	}
}

func (d *SheetValuesDataSource) Configure(
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

func (d *SheetValuesDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data SheetValuesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vr, err := d.client.Sheets.ValuesGet(ctx, data.SpreadsheetID.ValueString(), data.Range.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Sheet Values Failed", err.Error())
		return
	}

	rows, diags := valueRangeToRows(vr)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Rows = rows
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func valueRangeToRows(vr *sheets.ValueRange) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	rowType := types.ObjectType{AttrTypes: map[string]attr.Type{"cells": types.ListType{ElemType: types.StringType}}}
	out := make([]attr.Value, 0, len(vr.Values))

	for _, r := range vr.Values {
		cells := make([]attr.Value, 0, len(r))
		for _, c := range r {
			cells = append(cells, types.StringValue(fmt.Sprintf("%v", c)))
		}
		cellList, d := types.ListValue(types.StringType, cells)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(rowType), diags
		}

		obj, d := types.ObjectValue(rowType.AttrTypes, map[string]attr.Value{"cells": cellList})
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(rowType), diags
		}
		out = append(out, obj)
	}

	list, d := types.ListValue(rowType, out)
	diags.Append(d...)
	if diags.HasError() {
		return types.ListNull(rowType), diags
	}
	return list, diags
}
