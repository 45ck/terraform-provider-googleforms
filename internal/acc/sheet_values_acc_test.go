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

func TestAccSheetValues_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("skipping acceptance test unless TF_ACC is set")
	}
	if os.Getenv("GOOGLE_CREDENTIALS") == "" {
		t.Skip("skipping acceptance test unless GOOGLE_CREDENTIALS is set")
	}

	name := acctest.RandomWithPrefix("tf-test-googleforms-ss-values")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"googleforms": providerserver.NewProtocol6WithError(providerImpl.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: accSheetValuesConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("googleforms_sheet_values.test", "id"),
					resource.TestCheckResourceAttr("googleforms_sheet_values.test", "range", "A1:B2"),
				),
			},
			{
				ResourceName:      "googleforms_sheet_values.test",
				ImportState:       true,
				ImportStateVerify: false, // import only sets id/spreadsheet_id/range; values are not read-back by default.
			},
		},
	})
}

func accSheetValuesConfig(title string) string {
	return `
provider "googleforms" {}

resource "googleforms_spreadsheet" "test" {
  title = "` + title + `"
}

resource "googleforms_sheet_values" "test" {
  spreadsheet_id = googleforms_spreadsheet.test.id
  range          = "A1:B2"

  value_input_option = "RAW"
  read_back          = true

  rows = [
    { cells = ["a1", "b1"] },
    { cells = ["a2", "b2"] },
  ]
}
`
}
