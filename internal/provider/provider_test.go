package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories is a map that Terraform uses to load the provider during acceptance tests.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"qrcode": providerserver.NewProtocol6WithError(New("test")()),
}

func TestProvider(t *testing.T) {
	// This is a placeholder test to ensure the provider compiles and can be loaded.
	// Actual acceptance tests should be defined in separate test functions.
}
