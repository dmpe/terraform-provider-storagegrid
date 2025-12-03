// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestBucketDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"storagegrid": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: `data "storagegrid_bucket" "test" { name = "tf-provider-acc-test-bucket" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.storagegrid_bucket.test", "name", "tf-provider-acc-test-bucket"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket.test", "region", "us-east-1"),
					resource.TestCheckNoResourceAttr("data.storagegrid_bucket.test", "object_lock_configuration"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"storagegrid": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: `data "storagegrid_bucket" "test" { name = "tf-provider-acc-test-bucket-ol" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.storagegrid_bucket.test", "name", "tf-provider-acc-test-bucket-ol"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket.test", "region", "us-east-1"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket.test", "object_lock_configuration.mode", "governance"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket.test", "object_lock_configuration.days", "10"),
				),
			},
		},
	})
}
