// Copyright github.com/dmpe 2024, 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestS3KeysByUserIdResourceWithoutExpirationDate(t *testing.T) {
	username := fmt.Sprintf("tf-provider-acc-test-user-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"storagegrid": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Create
			{
				Config: fmt.Sprintf(`
resource "storagegrid_users" "test" {
	unique_name = "user/%s"
	full_name   = "%s"
	member_of   = []
}

resource "storagegrid_s3_access_key" "test" {
	user_uuid = storagegrid_users.test.id
}
`, username, username),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("storagegrid_s3_access_key.test", "user_uuid"),
					resource.TestCheckResourceAttr("storagegrid_s3_access_key.test", "expires", ""),
				),
			},
			// Delete testing is done automatically
		},
	})
}

func TestS3KeysByUserIdResourceWithExpirationDate(t *testing.T) {
	username := fmt.Sprintf("tf-provider-acc-test-user-%d", time.Now().Unix())
	expirationDate := time.Now().Add(10 * time.Minute).Format("2006-01-02T15:04:05Z")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"storagegrid": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Create
			{
				Config: fmt.Sprintf(`
resource "storagegrid_users" "test" {
	unique_name = "user/%s"
	full_name   = "%s"
	member_of   = []
}

resource "storagegrid_s3_access_key" "test" {
	user_uuid = storagegrid_users.test.id
	expires   = "%s"
}
`, username, username, expirationDate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("storagegrid_s3_access_key.test", "user_uuid"),
					resource.TestCheckResourceAttr("storagegrid_s3_access_key.test", "expires", expirationDate),
				),
			},
			// Delete testing is done automatically
		},
	})
}
