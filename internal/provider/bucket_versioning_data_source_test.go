// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestBucketVersioningDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"storagegrid": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: `data "storagegrid_bucket_versioning" "test" { bucket_name = "tf-provider-acc-test-bucket" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.storagegrid_bucket_versioning.test", "bucket_name", "tf-provider-acc-test-bucket"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_versioning.test", "status", "Disabled"),
				),
			},
		},
	})
}
