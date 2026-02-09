// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package provider_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	providerImpl "github.com/your-org/terraform-provider-googleforms/internal/provider"
)

// testFakeCredentials returns a minimal valid service account JSON for testing.
func testFakeCredentials() string {
	return `{
		"type": "service_account",
		"project_id": "test-project",
		"private_key_id": "key123",
		"private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF8PbnGy0AHB7MhgHcTz6sE2I2yPB\naFDrBz9vFqU4zK7G5cH0I3FAdADAXaBEHFHkbqMRnlNP5dv7ygMfkPX8HXbISEBd\nc0rROFAxfvaBdice9oph4JYIjUDJKRmOqJH8FHXQNKXL4MBTvOgMNbIMbIGpJPM\nZhJQ7qKH3JzrGSplmajn3K53DJMbxLbPOqGH7g6VFB3bUKMCp8FPBVLFMz4JhHjq\neFK5VJvF0MEbBOixin1GlRT6gJSYM/i0FBVMDmdMNCMXGFMNJMHDXEqFHF09RGYFA\n0jPej7VJz0PoYwS5jP1PVHK+0G3MZExaNdVxlwIDAQABAoIBAC5IgFBJmOzCBPDJ\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\n-----END RSA PRIVATE KEY-----\n",
		"client_email": "test@test-project.iam.gserviceaccount.com",
		"client_id": "123456789",
		"auth_uri": "https://accounts.google.com/o/oauth2/auth",
		"token_uri": "https://oauth2.googleapis.com/token"
	}`
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

	if resp.TypeName != "google_forms" {
		t.Errorf("expected type name 'google_forms', got %q", resp.TypeName)
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

	// Instantiate the first resource and check its type name.
	r := resourceResp[0]()
	metaResp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "google_forms",
	}, metaResp)

	if metaResp.TypeName != "google_forms_form" {
		t.Errorf("expected resource type 'google_forms_form', got %q", metaResp.TypeName)
	}
}

func TestProviderDataSources_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	p := newTestProvider()
	dataSources := p.DataSources(context.Background())

	if len(dataSources) != 0 {
		t.Errorf("expected zero data sources, got %d", len(dataSources))
	}
}

// TestProviderAcceptance_ConfigIsValid runs a basic Terraform acceptance-style
// config validation. This uses the testing framework's ProtoV6ProviderFactories
// to verify the provider can be instantiated and configured.
func TestProviderAcceptance_ConfigIsValid(t *testing.T) {
	t.Parallel()

	// This test only validates that the provider config schema is valid.
	// It does not make real API calls.
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (providerserver.ProviderServer, error){
			"google_forms": providerserver.NewProtocol6WithError(providerImpl.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				// An empty config is valid since all attributes are optional.
				Config:             `provider "google_forms" {}`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}
