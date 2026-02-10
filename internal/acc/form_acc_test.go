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

func TestAccForm_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("skipping acceptance test unless TF_ACC is set")
	}
	if os.Getenv("GOOGLE_CREDENTIALS") == "" {
		t.Skip("skipping acceptance test unless GOOGLE_CREDENTIALS is set")
	}

	title := acctest.RandomWithPrefix("tf-test-googleforms-form")
	folder := acctest.RandomWithPrefix("tf-test-googleforms-form-folder")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"googleforms": providerserver.NewProtocol6WithError(providerImpl.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: accFormConfig(title, folder),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googleforms_form.test", "title", title),
					resource.TestCheckResourceAttrSet("googleforms_form.test", "id"),
					resource.TestCheckResourceAttrSet("googleforms_form.test", "responder_uri"),
				),
			},
			{
				ResourceName:      "googleforms_form.test",
				ImportState:       true,
				ImportStateVerify: false, // item_key values are auto-generated during import.
			},
		},
	})
}

func accFormConfig(title string, folderName string) string {
	return `
provider "googleforms" {}

resource "googleforms_drive_folder" "test" {
  name = "` + folderName + `"
}

resource "googleforms_form" "test" {
  title       = "` + title + `"
  description = "Acceptance test"

  folder_id = googleforms_drive_folder.test.id

  manage_mode     = "all"
  update_strategy = "replace_all"

  item {
    item_key = "name"
    short_answer {
      question_text = "What is your name?"
      required      = true
    }
  }
}
`
}
