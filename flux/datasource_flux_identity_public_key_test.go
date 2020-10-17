package flux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFluxIdentity(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccFluxIdentityBasic(),
				Check: resource.ComposeTestCheckFunc(
					testDataSourceReturned("data.flux_identity_public_key.identity"),
				),
			},
		},
	})
}

func testAccFluxIdentityBasic() string {
	return fmt.Sprintf(`
	data "flux_identity_public_key" "identity" {
		namespace = "flux"
		pod_labels = {
    		app = "flux"
        }
	}
	`)
}

func testDataSourceReturned(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.Attributes["public_key"] == "" {
			return fmt.Errorf("no public_key set")
		}

		return nil
	}
}

func providerFactories() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"flux": func() (*schema.Provider, error) {
			return Provider(), nil
		},
	}
}
