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

func TestBucketQuotaResource(t *testing.T) {
	bucketName := fmt.Sprintf("tf-provider-acc-test-bucket-quota-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"storagegrid": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Create
			{
				Config: bucketQuotaConfiguration(bucketName, 1000000000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("storagegrid_bucket_quota.test", "bucket_name", bucketName),
					resource.TestCheckResourceAttr("storagegrid_bucket_quota.test", "object_bytes", "1000000000"),
				),
			},
			// Update
			{
				Config: bucketQuotaConfiguration(bucketName, 2000000000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("storagegrid_bucket_quota.test", "bucket_name", bucketName),
					resource.TestCheckResourceAttr("storagegrid_bucket_quota.test", "object_bytes", "2000000000"),
				),
			},
			// Import
			{
				ResourceName:                         "storagegrid_bucket_quota.test",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "bucketName",
				ImportStateId:                        bucketName,
			},
			// Delete testing is done automatically
		},
	})
}

func bucketQuotaConfiguration(bucketName string, quota int64) string {
	bucketResource := fmt.Sprintf(`
resource "storagegrid_bucket" "test" {
	name = "%s"
}`, bucketName)

	versioningConfiguration := fmt.Sprintf(`
resource "storagegrid_bucket_quota" "test" {
	bucket_name = storagegrid_bucket.test.name
	object_bytes = "%d"
}`, quota)

	return fmt.Sprintf("%s\n\n%s", bucketResource, versioningConfiguration)
}
