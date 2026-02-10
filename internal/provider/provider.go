// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

// Package provider implements the Terraform provider for Google Forms.
package provider

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/45ck/terraform-provider-googleforms/internal/client"
	datasourcesheetvalues "github.com/45ck/terraform-provider-googleforms/internal/datasource_sheet_values"
	datasourcespreadsheet "github.com/45ck/terraform-provider-googleforms/internal/datasource_spreadsheet"
	resourcedrivefile "github.com/45ck/terraform-provider-googleforms/internal/resource_drive_file"
	resourcedrivefolder "github.com/45ck/terraform-provider-googleforms/internal/resource_drive_folder"
	resourcedrivepermission "github.com/45ck/terraform-provider-googleforms/internal/resource_drive_permission"
	resourceform "github.com/45ck/terraform-provider-googleforms/internal/resource_form"
	resourceformsbatchupdate "github.com/45ck/terraform-provider-googleforms/internal/resource_forms_batch_update"
	resourceresponsesheet "github.com/45ck/terraform-provider-googleforms/internal/resource_response_sheet"
	resourcesheet "github.com/45ck/terraform-provider-googleforms/internal/resource_sheet"
	resourcesheetvalues "github.com/45ck/terraform-provider-googleforms/internal/resource_sheet_values"
	resourcesheetsbatchupdate "github.com/45ck/terraform-provider-googleforms/internal/resource_sheets_batch_update"
	resourcesheetsconditionalformatrule "github.com/45ck/terraform-provider-googleforms/internal/resource_sheets_conditional_format_rule"
	resourcesheetsdatavalidation "github.com/45ck/terraform-provider-googleforms/internal/resource_sheets_data_validation"
	resourcesheetsdevelopermetadata "github.com/45ck/terraform-provider-googleforms/internal/resource_sheets_developer_metadata"
	resourcesheetsnamedrange "github.com/45ck/terraform-provider-googleforms/internal/resource_sheets_named_range"
	resourcesheetsprotectedrange "github.com/45ck/terraform-provider-googleforms/internal/resource_sheets_protected_range"
	resourcespreadsheet "github.com/45ck/terraform-provider-googleforms/internal/resource_spreadsheet"
)

var _ provider.Provider = &GoogleFormsProvider{}

// GoogleFormsProvider implements the Terraform provider for Google Forms.
type GoogleFormsProvider struct {
	version string
}

// GoogleFormsProviderModel describes the provider configuration data.
type GoogleFormsProviderModel struct {
	Credentials     types.String `tfsdk:"credentials"`
	ImpersonateUser types.String `tfsdk:"impersonate_user"`
}

// New returns a new provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GoogleFormsProvider{
			version: version,
		}
	}
}

func (p *GoogleFormsProvider) Metadata(
	_ context.Context,
	_ provider.MetadataRequest,
	resp *provider.MetadataResponse,
) {
	resp.TypeName = "googleforms"
	resp.Version = p.version
}

func (p *GoogleFormsProvider) Schema(
	_ context.Context,
	_ provider.SchemaRequest,
	resp *provider.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manage Google Forms as infrastructure using Terraform.",
		Attributes: map[string]schema.Attribute{
			"credentials": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Description: "Service account JSON key or path to a JSON key file. " +
					"Falls back to GOOGLE_CREDENTIALS env var, then Application Default Credentials.",
			},
			"impersonate_user": schema.StringAttribute{
				Optional:    true,
				Description: "Email of user to impersonate via domain-wide delegation.",
			},
		},
	}
}

func (p *GoogleFormsProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	var config GoogleFormsProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	credentialsJSON := resolveCredentials(ctx, config.Credentials)

	var impersonateUser string
	if !config.ImpersonateUser.IsNull() && !config.ImpersonateUser.IsUnknown() {
		impersonateUser = config.ImpersonateUser.ValueString()
	}

	tflog.Debug(ctx, "creating Google Forms API client",
		map[string]interface{}{
			"has_credentials":  credentialsJSON != "",
			"impersonate_user": impersonateUser,
		},
	)

	apiClient, err := client.NewClient(ctx, credentialsJSON, impersonateUser)
	if err != nil {
		resp.Diagnostics.AddError("Client Creation Failed",
			"Unable to create Google Forms API client: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient
}

// resolveCredentials determines the credentials JSON string from the provider
// config value, the GOOGLE_CREDENTIALS environment variable, or returns empty
// to signal that Application Default Credentials should be used.
func resolveCredentials(ctx context.Context, attr types.String) string {
	// 1. Explicit provider config value takes priority.
	if !attr.IsNull() && !attr.IsUnknown() {
		return resolveCredentialValue(ctx, attr.ValueString())
	}

	// 2. Fall back to the GOOGLE_CREDENTIALS environment variable.
	if envVal := os.Getenv("GOOGLE_CREDENTIALS"); envVal != "" {
		tflog.Info(ctx, "using credentials from GOOGLE_CREDENTIALS environment variable")
		return resolveCredentialValue(ctx, envVal)
	}

	// 3. No explicit credentials -- ADC will be used by the client.
	tflog.Info(ctx, "no explicit credentials; falling back to Application Default Credentials")
	return ""
}

// resolveCredentialValue checks whether the value looks like a file path or
// raw JSON. If it looks like a file path and the file exists, its contents are
// returned. Otherwise the value is returned as-is (assumed to be JSON).
func resolveCredentialValue(ctx context.Context, value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	// If the value starts with '{', treat it as inline JSON.
	if strings.HasPrefix(trimmed, "{") {
		return trimmed
	}

	// Otherwise treat it as a file path.
	data, err := os.ReadFile(trimmed)
	if err != nil {
		tflog.Warn(ctx, "credentials value is not valid JSON and could not be read as a file; using as-is",
			map[string]interface{}{"error": err.Error()},
		)
		return trimmed
	}

	tflog.Debug(ctx, "loaded credentials from file", map[string]interface{}{"path": trimmed})
	return string(data)
}

func (p *GoogleFormsProvider) Resources(
	_ context.Context,
) []func() resource.Resource {
	return []func() resource.Resource{
		resourceform.NewFormResource,
		resourceformsbatchupdate.NewFormsBatchUpdateResource,
		resourcespreadsheet.NewSpreadsheetResource,
		resourcesheet.NewSheetResource,
		resourcesheetvalues.NewSheetValuesResource,
		resourcesheetsbatchupdate.NewSheetsBatchUpdateResource,
		resourcesheetsnamedrange.NewNamedRangeResource,
		resourcesheetsprotectedrange.NewProtectedRangeResource,
		resourcesheetsdevelopermetadata.NewDeveloperMetadataResource,
		resourcesheetsdatavalidation.NewDataValidationResource,
		resourcesheetsconditionalformatrule.NewConditionalFormatRuleResource,
		resourcedrivefolder.NewDriveFolderResource,
		resourcedrivefile.NewDriveFileResource,
		resourcedrivepermission.NewDrivePermissionResource,
		resourceresponsesheet.NewResponseSheetResource,
	}
}

func (p *GoogleFormsProvider) DataSources(
	_ context.Context,
) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasourcespreadsheet.NewSpreadsheetDataSource,
		datasourcesheetvalues.NewSheetValuesDataSource,
	}
}
