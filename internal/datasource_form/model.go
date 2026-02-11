// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package datasourceform

import "github.com/hashicorp/terraform-plugin-framework/types"

type FormDataSourceModel struct {
	ID types.String `tfsdk:"id"`

	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`

	ResponderURI  types.String `tfsdk:"responder_uri"`
	EditURI       types.String `tfsdk:"edit_uri"`
	DocumentTitle types.String `tfsdk:"document_title"`
	RevisionID    types.String `tfsdk:"revision_id"`

	LinkedSheetID       types.String `tfsdk:"linked_sheet_id"`
	Quiz                types.Bool   `tfsdk:"quiz"`
	EmailCollectionType types.String `tfsdk:"email_collection_type"`
}
