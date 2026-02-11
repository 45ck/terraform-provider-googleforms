// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package resourceform

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
)

// Compile-time interface checks.
var (
	_ resource.Resource                     = &FormResource{}
	_ resource.ResourceWithImportState      = &FormResource{}
	_ resource.ResourceWithConfigValidators = &FormResource{}
)

// FormResource implements the googleforms_form Terraform resource.
type FormResource struct {
	client *client.Client
}

// NewFormResource returns a new resource factory function.
func NewFormResource() resource.Resource {
	return &FormResource{}
}

// Metadata sets the resource type name.
func (r *FormResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_form"
}

// Configure extracts the provider-configured client.
func (r *FormResource) Configure(
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

// ConfigValidators returns all resource-level config validators.
func (r *FormResource) ConfigValidators(
	_ context.Context,
) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		MutuallyExclusiveValidator{},
		AcceptingResponsesRequiresPublishedValidator{},
		UniqueItemKeyValidator{},
		ExactlyOneSubBlockValidator{},
		OptionsRequiredForChoiceValidator{},
		ChoiceOptionNavigationValidator{},
		CorrectAnswerInOptionsValidator{},
		GradingRequiresQuizValidator{},
	}
}
