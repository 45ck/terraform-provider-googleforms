// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

// Package resourceformsbatchupdate implements the googleforms_forms_batch_update Terraform resource.
package resourceformsbatchupdate

import "github.com/hashicorp/terraform-plugin-framework/types"

// FormsBatchUpdateResourceModel describes the Terraform state for googleforms_forms_batch_update.
type FormsBatchUpdateResourceModel struct {
	ID types.String `tfsdk:"id"`

	FormID types.String `tfsdk:"form_id"`

	RequestsJSON          types.String `tfsdk:"requests_json"`
	IncludeFormInResponse types.Bool   `tfsdk:"include_form_in_response"`
	StoreResponseJSON     types.Bool   `tfsdk:"store_response_json"`
	RequiredRevisionID    types.String `tfsdk:"required_revision_id"`
	ResponseJSON          types.String `tfsdk:"response_json"`
}
