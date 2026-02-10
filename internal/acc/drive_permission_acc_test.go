// Copyright 2026 terraform-provider-googleforms contributors
// SPDX-License-Identifier: Apache-2.0

package acc

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	providerImpl "github.com/45ck/terraform-provider-googleforms/internal/provider"
)

func TestAccDrivePermission_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("skipping acceptance test unless TF_ACC is set")
	}
	if os.Getenv("GOOGLE_CREDENTIALS") == "" {
		t.Skip("skipping acceptance test unless GOOGLE_CREDENTIALS is set")
	}

	grantee := os.Getenv("GOOGLEFORMS_TEST_GRANTEE_EMAIL")
	if grantee == "" {
		t.Skip("skipping drive_permission acceptance test unless GOOGLEFORMS_TEST_GRANTEE_EMAIL is set")
	}

	name := acctest.RandomWithPrefix("tf-test-googleforms-perm-ss")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"googleforms": providerserver.NewProtocol6WithError(providerImpl.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: accDrivePermissionConfig(name, grantee),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("googleforms_drive_permission.test", "id"),
					resource.TestCheckResourceAttr("googleforms_drive_permission.test", "type", "user"),
					resource.TestCheckResourceAttr("googleforms_drive_permission.test", "role", "reader"),
				),
			},
		},
	})
}

func accDrivePermissionConfig(title string, email string) string {
	return `
provider "googleforms" {}

resource "googleforms_spreadsheet" "test" {
  title = "` + title + `"
}

resource "googleforms_drive_permission" "test" {
  file_id       = googleforms_spreadsheet.test.id
  type          = "user"
  role          = "reader"
  email_address = "` + email + `"

  send_notification_email = false
}
`
}
