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

func TestBucketVersioningResource(t *testing.T) {
	bucketName := fmt.Sprintf("tf-provider-acc-test-bucket-versioning-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"storagegrid": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Create
			{
				Config: bucketVersioningConfiguration(bucketName, "Enabled"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("storagegrid_bucket_versioning.test", "bucket_name", bucketName),
					resource.TestCheckResourceAttr("storagegrid_bucket_versioning.test", "status", "Enabled"),
				),
			},
			// Update
			{
				Config: bucketVersioningConfiguration(bucketName, "Suspended"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("storagegrid_bucket_versioning.test", "bucket_name", bucketName),
					resource.TestCheckResourceAttr("storagegrid_bucket_versioning.test", "status", "Suspended"),
				),
			},
			// Import
			{
				ResourceName:                         "storagegrid_bucket_versioning.test",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "bucketName",
				ImportStateId:                        bucketName,
			},
			// Delete testing is done automatically
		},
	})
}

func bucketVersioningConfiguration(bucketName string, status string) string {
	bucketResource := fmt.Sprintf(`
resource "storagegrid_bucket" "test" {
	name = "%s"
}`, bucketName)

	versioningConfiguration := fmt.Sprintf(`
resource "storagegrid_bucket_versioning" "test" {
	bucket_name = storagegrid_bucket.test.name
	status = "%s"
}`, status)

	return fmt.Sprintf("%s\n\n%s", bucketResource, versioningConfiguration)
}
