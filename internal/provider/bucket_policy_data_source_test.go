// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestBucketPolicyDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"storagegrid": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: `data "storagegrid_bucket_policy" "test" { bucket_name = "tf-provider-acc-test-bucket" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "bucket_name", "tf-provider-acc-test-bucket"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.id", ""),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.version", ""),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.#", "2"),
					// First Statement
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.0.sid", "test-sid-1"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.0.effect", "Allow"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.0.action.#", "1"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.0.action.0", "s3:ListBucket"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.0.resource.#", "2"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.0.resource.0", "arn:aws:s3:::tf-provider-acc-test-bucket"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.0.resource.1", "arn:aws:s3:::tf-provider-acc-test-bucket/*"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.0.principal.type", "*"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.0.principal.identifiers.#", "0"),
					resource.TestCheckNoResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.0.condition"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.0.not_resource.#", "0"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.0.not_action.#", "0"),
					resource.TestCheckNoResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.0.not_principal"),
					// Second Statement
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.1.sid", "test-sid-2"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.1.effect", "Allow"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.1.action.#", "2"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.1.action.0", "s3:ListBucket"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.1.action.1", "s3:GetObject"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.1.resource.#", "2"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.1.resource.0", "arn:aws:s3:::tf-provider-acc-test-bucket"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.1.resource.1", "arn:aws:s3:::tf-provider-acc-test-bucket/*"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.1.principal.type", "*"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.1.principal.identifiers.#", "0"),
					resource.TestCheckNoResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.1.condition"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.1.not_resource.#", "0"),
					resource.TestCheckResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.1.not_action.#", "0"),
					resource.TestCheckNoResourceAttr("data.storagegrid_bucket_policy.test", "policy.statement.1.not_principal"),
				),
			},
		},
	})
}
