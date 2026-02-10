// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package provider_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	providerImpl "github.com/45ck/terraform-provider-googleforms/internal/provider"
)

// testFakeCredentials returns a minimal valid service account JSON for testing.
func testFakeCredentials() string {
	// Intentionally not a real PEM key (keeps tests offline and avoids gosec
	// false positives for hardcoded credentials).
	privateKey := "not-a-real-private-key"

	// Keep this structurally valid (json.Marshal will escape newlines).
	creds := map[string]any{
		"type":           "service_account",
		"project_id":     "test-project",
		"private_key_id": "key123",
		"private_key":    privateKey,
		"client_email":   "test@test-project.iam.gserviceaccount.com",
		"client_id":      "123456789",
		"auth_uri":       "https://accounts.google.com/o/oauth2/auth",
		"token_uri":      "https://oauth2.googleapis.com/token",
	}

	b, err := json.Marshal(creds)
	if err != nil {
		// Should never happen with the static test map above.
		panic(err)
	}
	return string(b)
}

// newTestProvider creates a GoogleFormsProvider via the New factory for testing.
func newTestProvider() provider.Provider {
	return providerImpl.New("test")()
}

func TestProviderSchema_HasRequiredAttributes(t *testing.T) {
	t.Parallel()

	p := newTestProvider()
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %s", resp.Diagnostics)
	}

	attrs := resp.Schema.Attributes
	if _, ok := attrs["credentials"]; !ok {
		t.Error("schema missing 'credentials' attribute")
	}
	if _, ok := attrs["impersonate_user"]; !ok {
		t.Error("schema missing 'impersonate_user' attribute")
	}
}

func TestProviderSchema_CredentialsIsSensitive(t *testing.T) {
	t.Parallel()

	p := newTestProvider()
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %s", resp.Diagnostics)
	}

	credAttr := resp.Schema.Attributes["credentials"]
	if !credAttr.IsSensitive() {
		t.Error("credentials attribute should be marked as sensitive")
	}
}

func TestProviderConfigure_WithCredentials(t *testing.T) {
	t.Parallel()

	p := newTestProvider()

	// Build a tftypes.Value representing the provider config with credentials set.
	configVal := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"credentials":      tftypes.String,
			"impersonate_user": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"credentials":      tftypes.NewValue(tftypes.String, testFakeCredentials()),
		"impersonate_user": tftypes.NewValue(tftypes.String, nil),
	})

	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, schemaResp)

	configState, err := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    configVal,
	}, error(nil)
	if err != nil {
		t.Fatalf("failed to create config: %s", err)
	}

	resp := &provider.ConfigureResponse{}
	p.Configure(context.Background(), provider.ConfigureRequest{
		Config: configState,
	}, resp)

	// The Configure call will fail because client.NewClient is a TODO stub,
	// but we verify it attempted to create the client (error should mention
	// client initialization, not a credentials-missing error).
	if !resp.Diagnostics.HasError() {
		// If no error, the client was created successfully (future implementation).
		return
	}

	// Accept the "not yet implemented" error from the stub NewClient.
	for _, d := range resp.Diagnostics.Errors() {
		if d.Summary() == "Client Creation Failed" {
			return
		}
	}
	t.Errorf("expected client creation error, got: %s", resp.Diagnostics)
}

func TestProviderConfigure_FallbackToEnvVar(t *testing.T) {
	// Not parallel: modifies environment variables.
	creds := testFakeCredentials()

	// Write credentials to a temp file to test file-path resolution.
	tmpDir := t.TempDir()
	credFile := filepath.Join(tmpDir, "creds.json")
	if err := os.WriteFile(credFile, []byte(creds), 0600); err != nil {
		t.Fatalf("failed to write temp cred file: %s", err)
	}

	t.Setenv("GOOGLE_CREDENTIALS", credFile)

	p := newTestProvider()

	// Config with null credentials (not set by user).
	configVal := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"credentials":      tftypes.String,
			"impersonate_user": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"credentials":      tftypes.NewValue(tftypes.String, nil),
		"impersonate_user": tftypes.NewValue(tftypes.String, nil),
	})

	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, schemaResp)

	resp := &provider.ConfigureResponse{}
	p.Configure(context.Background(), provider.ConfigureRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    configVal,
		},
	}, resp)

	// The stub NewClient returns an error, but the important thing is that
	// Configure did not add an error about missing credentials -- it
	// proceeded to call NewClient.
	if !resp.Diagnostics.HasError() {
		return
	}
	for _, d := range resp.Diagnostics.Errors() {
		if d.Summary() == "Client Creation Failed" {
			return
		}
	}
	t.Errorf("expected client creation error (env fallback), got: %s", resp.Diagnostics)
}

func TestProviderConfigure_FallbackToADC(t *testing.T) {
	// Not parallel: modifies environment variables.
	t.Setenv("GOOGLE_CREDENTIALS", "")

	p := newTestProvider()

	// Config with null credentials and no env var.
	configVal := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"credentials":      tftypes.String,
			"impersonate_user": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"credentials":      tftypes.NewValue(tftypes.String, nil),
		"impersonate_user": tftypes.NewValue(tftypes.String, nil),
	})

	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, schemaResp)

	resp := &provider.ConfigureResponse{}
	p.Configure(context.Background(), provider.ConfigureRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    configVal,
		},
	}, resp)

	// With no credentials at all, Configure should still call NewClient
	// with empty credentials (ADC path). The stub will return an error.
	if !resp.Diagnostics.HasError() {
		return
	}
	for _, d := range resp.Diagnostics.Errors() {
		if d.Summary() == "Client Creation Failed" {
			return
		}
	}
	t.Errorf("expected client creation error (ADC fallback), got: %s", resp.Diagnostics)
}

func TestProviderConfigure_InvalidCredentials(t *testing.T) {
	t.Parallel()

	p := newTestProvider()

	// Provide obviously invalid credentials JSON.
	configVal := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"credentials":      tftypes.String,
			"impersonate_user": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"credentials":      tftypes.NewValue(tftypes.String, "not-valid-json"),
		"impersonate_user": tftypes.NewValue(tftypes.String, nil),
	})

	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, schemaResp)

	resp := &provider.ConfigureResponse{}
	p.Configure(context.Background(), provider.ConfigureRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    configVal,
		},
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostic for invalid credentials, got none")
	}
}

func TestProviderMetadata_TypeName(t *testing.T) {
	t.Parallel()

	p := newTestProvider()
	resp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)

	if resp.TypeName != "googleforms" {
		t.Errorf("expected type name 'googleforms', got %q", resp.TypeName)
	}
}

func TestProviderMetadata_Version(t *testing.T) {
	t.Parallel()

	p := providerImpl.New("1.2.3")()
	resp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)

	if resp.Version != "1.2.3" {
		t.Errorf("expected version '1.2.3', got %q", resp.Version)
	}
}

func TestProviderResources_RegistersFormResource(t *testing.T) {
	t.Parallel()

	p := newTestProvider()

	resourceResp := p.Resources(context.Background())
	if len(resourceResp) == 0 {
		t.Fatal("expected at least one resource factory, got none")
	}

	// Instantiate all resources and ensure the expected ones are registered.
	want := map[string]bool{
		"googleforms_form":                           false,
		"googleforms_response_sheet":                 false,
		"googleforms_spreadsheet":                    false,
		"googleforms_sheet":                          false,
		"googleforms_sheet_values":                   false,
		"googleforms_sheets_batch_update":            false,
		"googleforms_sheets_named_range":             false,
		"googleforms_sheets_protected_range":         false,
		"googleforms_sheets_developer_metadata":      false,
		"googleforms_sheets_data_validation":         false,
		"googleforms_sheets_conditional_format_rule": false,
		"googleforms_drive_permission":               false,
	}

	for _, f := range resourceResp {
		r := f()
		metaResp := &fwresource.MetadataResponse{}
		r.Metadata(context.Background(), fwresource.MetadataRequest{ProviderTypeName: "googleforms"}, metaResp)
		if _, ok := want[metaResp.TypeName]; ok {
			want[metaResp.TypeName] = true
		}
	}

	for typ, seen := range want {
		if !seen {
			t.Errorf("expected resource type %q to be registered", typ)
		}
	}
}

func TestProviderDataSources_RegistersExpectedDataSources(t *testing.T) {
	t.Parallel()

	p := newTestProvider()
	dataSources := p.DataSources(context.Background())

	if len(dataSources) == 0 {
		t.Fatal("expected at least one data source factory, got none")
	}

	want := map[string]bool{
		"googleforms_spreadsheet":  false,
		"googleforms_sheet_values": false,
	}

	for _, f := range dataSources {
		ds := f()
		metaResp := &datasource.MetadataResponse{}
		ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "googleforms"}, metaResp)
		if _, ok := want[metaResp.TypeName]; ok {
			want[metaResp.TypeName] = true
		}
	}

	for typ, seen := range want {
		if !seen {
			t.Errorf("expected data source type %q to be registered", typ)
		}
	}
}

// TestProviderAcceptance_ConfigIsValid runs a basic Terraform acceptance-style
// config validation. This uses the testing framework's ProtoV6ProviderFactories
// to verify the provider can be instantiated and configured.
func TestProviderAcceptance_ConfigIsValid(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("skipping acceptance-style config validation test unless TF_ACC is set")
	}

	// This test only validates that the provider config schema is valid.
	// It does not make real API calls.
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"googleforms": providerserver.NewProtocol6WithError(providerImpl.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				// An empty config is valid since all attributes are optional.
				Config:             `provider "googleforms" {}`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}
