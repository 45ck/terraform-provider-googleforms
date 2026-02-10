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

func TestAccSpreadsheet_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("skipping acceptance test unless TF_ACC is set")
	}
	if os.Getenv("GOOGLE_CREDENTIALS") == "" {
		t.Skip("skipping acceptance test unless GOOGLE_CREDENTIALS is set")
	}

	name := acctest.RandomWithPrefix("tf-test-googleforms-ss")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"googleforms": providerserver.NewProtocol6WithError(providerImpl.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: accSpreadsheetConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googleforms_spreadsheet.test", "title", name),
					resource.TestCheckResourceAttrSet("googleforms_spreadsheet.test", "id"),
					resource.TestCheckResourceAttrSet("googleforms_spreadsheet.test", "url"),
				),
			},
			{
				ResourceName:      "googleforms_spreadsheet.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func accSpreadsheetConfig(title string) string {
	return `
provider "googleforms" {}

resource "googleforms_spreadsheet" "test" {
  title = "` + title + `"
}
`
}
