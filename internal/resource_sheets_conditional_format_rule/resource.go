// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourcesheetsconditionalformatrule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

var (
	_ resource.Resource                = &ConditionalFormatRuleResource{}
	_ resource.ResourceWithImportState = &ConditionalFormatRuleResource{}
)

// ConditionalFormatRuleResource implements googleforms_sheets_conditional_format_rule.
type ConditionalFormatRuleResource struct {
	client *client.Client
}

func NewConditionalFormatRuleResource() resource.Resource {
	return &ConditionalFormatRuleResource{}
}

func (r *ConditionalFormatRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sheets_conditional_format_rule"
}

func (r *ConditionalFormatRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *client.Client, got unexpected type.")
		return
	}
	r.client = c
}
