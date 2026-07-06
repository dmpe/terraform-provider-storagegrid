// Copyright github.com/dmpe 2024, 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestFederatedUsersResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	federatedUser := os.Getenv("STORAGEGRID_TEST_FEDERATED_USER")
	storageGridVersion, err := strconv.Atoi(os.Getenv("STORAGEGRID_TEST_GRID_VERSION"))
	if err != nil {
		t.Fatalf("Failed to convert STORAGEGRID_TEST_GRID_VERSION to int: %v", err)
	}

	var steps []resource.TestStep
	if storageGridVersion >= 12 {
		steps = append(steps, []resource.TestStep{
			// Create - error 404 user not found
			{
				Config: `
resource "storagegrid_federated_users" "test" {
	unique_name = "tf-provider-acc-test-federated-user"
}`,
				ExpectError: regexp.MustCompile("404"),
			},
			// Create
			{
				Config: fmt.Sprintf(`
resource "storagegrid_federated_users" "test" {
	unique_name = "federated-user/%s"
}`, federatedUser),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("storagegrid_federated_users.test", "unique_name", fmt.Sprintf("federated-user/%s", federatedUser)),
					resource.TestCheckResourceAttrSet("storagegrid_federated_users.test", "id"),
					resource.TestCheckResourceAttrSet("storagegrid_federated_users.test", "full_name"),
					resource.TestCheckResourceAttrSet("storagegrid_federated_users.test", "member_of.#"),
					resource.TestCheckResourceAttrSet("storagegrid_federated_users.test", "user_urn"),
					resource.TestCheckResourceAttrSet("storagegrid_federated_users.test", "account_id"),
				),
			},
		}...)
	} else {
		steps = append(steps, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "storagegrid_federated_users" "test" {
	unique_name = "federated-user/%s"
}`, federatedUser),
				ExpectError: regexp.MustCompile("Unsupported Product Version"),
			},
		}...)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"storagegrid": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: steps,
	})
}
