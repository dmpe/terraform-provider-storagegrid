// Copyright (c) github.com/dmpe
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

func TestBucketResource_DefaultRegion(t *testing.T) {
	bucketName := fmt.Sprintf("tf-provider-acc-test-bucket-%d", time.Now().Unix())
	defaultRegion := "us-east-1"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"storagegrid": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Create
			{
				Config: bucketResourceWithRegion(bucketName, nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("storagegrid_bucket.test", "name", bucketName),
					resource.TestCheckResourceAttr("storagegrid_bucket.test", "region", defaultRegion),
				),
			},
			// Import
			{
				ResourceName:                         "storagegrid_bucket.test",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateId:                        bucketName,
			},
			// Delete testing is done automatically
		},
	})
}

func TestBucketResource_CustomRegion(t *testing.T) {
	bucketName := fmt.Sprintf("tf-provider-acc-test-bucket-%d", time.Now().Unix())
	region := "eu-west-1"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"storagegrid": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Create
			{
				Config: bucketResourceWithRegion(bucketName, &region),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("storagegrid_bucket.test", "name", bucketName),
					resource.TestCheckResourceAttr("storagegrid_bucket.test", "region", region),
				),
			},
			// Import
			{
				ResourceName:                         "storagegrid_bucket.test",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateId:                        bucketName,
			},
			// Delete testing is done automatically
		},
	})
}

func bucketResourceWithRegion(name string, region *string) string {
	if region != nil {
		return fmt.Sprintf(`
resource "storagegrid_bucket" "test" {
	name = "%s"
	region = "%s"
}
`, name, *region)
	}

	return fmt.Sprintf(`
resource "storagegrid_bucket" "test" {
	name = "%s"
}
`, name)
}
