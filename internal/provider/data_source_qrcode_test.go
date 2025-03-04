package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccQRCodeDataSource verifies the qrcode_generate data source.
func TestAccQRCodeDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "qrcode" {}

					data "qrcode_generate" "test" {
						text = "qrcode"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify that the 'ascii' attribute is set
					resource.TestCheckResourceAttrSet(
						"data.qrcode_generate.test", "ascii",
					),
					// Optionally, verify that the 'ascii' attribute contains expected patterns
					resource.TestCheckResourceAttr(
						"data.qrcode_generate.test", "ascii_sha256",
						"1008c2f94d40f67e0f9f212284e9535aff2919fb256d512ad5edfa02929b55a5",
					),
				),
			},
		},
	})
}
