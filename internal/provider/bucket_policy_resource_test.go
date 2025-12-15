// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestBucketPolicyResource(t *testing.T) {
	bucketName := fmt.Sprintf("tf-provider-acc-test-bucket-policy-%d", time.Now().Unix())

	tenantId := os.Getenv("STORAGEGRID_TENANT")
	rootUserArn := fmt.Sprintf("arn:aws:iam::%s:root", tenantId)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"storagegrid": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Create
			{
				Config: bucketPolicyConfiguration(bucketName, "*", nil, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "bucket_name", bucketName),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.id", ""),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.version", ""),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.#", "1"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.sid", "test-sid"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.effect", "Allow"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.action.#", "1"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.action.0", "s3:ListBucket"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.resource.#", "2"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.resource.0", fmt.Sprintf("arn:aws:s3:::%s", bucketName)),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.resource.1", fmt.Sprintf("arn:aws:s3:::%s/*", bucketName)),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.principal.type", "*"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.principal.identifiers.#", "0"),
					resource.TestCheckNoResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.condition"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.not_resource.#", "0"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.not_action.#", "0"),
					resource.TestCheckNoResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.not_principal"),
				),
			},
			// Import
			{
				ResourceName:                         "storagegrid_bucket_policy.test",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "bucketName",
				ImportStateId:                        bucketName,
			},
			// Update Principal to Wildcard AWS
			{
				Config: bucketPolicyConfiguration(bucketName, "AWS", nil, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "bucket_name", bucketName),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.id", ""),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.version", ""),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.#", "1"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.sid", "test-sid"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.effect", "Allow"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.action.#", "1"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.action.0", "s3:ListBucket"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.resource.#", "2"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.resource.0", fmt.Sprintf("arn:aws:s3:::%s", bucketName)),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.resource.1", fmt.Sprintf("arn:aws:s3:::%s/*", bucketName)),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.principal.type", "AWS"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.principal.identifiers.#", "0"),
					resource.TestCheckNoResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.condition"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.not_resource.#", "0"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.not_action.#", "0"),
					resource.TestCheckNoResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.not_principal"),
				),
			},
			// Update Principal to principal identifiers on AWS
			{
				Config: bucketPolicyConfiguration(bucketName, "AWS", []string{fmt.Sprintf("\"%s\"", rootUserArn)}, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "bucket_name", bucketName),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.id", ""),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.version", ""),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.#", "1"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.sid", "test-sid"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.effect", "Allow"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.action.#", "1"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.action.0", "s3:ListBucket"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.resource.#", "2"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.resource.0", fmt.Sprintf("arn:aws:s3:::%s", bucketName)),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.resource.1", fmt.Sprintf("arn:aws:s3:::%s/*", bucketName)),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.principal.type", "AWS"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.principal.identifiers.#", "1"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.principal.identifiers.0", rootUserArn),
					resource.TestCheckNoResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.condition"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.not_resource.#", "0"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.not_action.#", "0"),
					resource.TestCheckNoResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.not_principal"),
				),
			},
			// Update Statement to use condition
			{
				Config: bucketPolicyConfiguration(bucketName, "AWS", []string{fmt.Sprintf("\"%s\"", rootUserArn)}, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "bucket_name", bucketName),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.id", ""),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.version", ""),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.#", "1"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.sid", "test-sid"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.effect", "Allow"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.action.#", "1"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.action.0", "s3:ListBucket"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.resource.#", "2"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.resource.0", fmt.Sprintf("arn:aws:s3:::%s", bucketName)),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.resource.1", fmt.Sprintf("arn:aws:s3:::%s/*", bucketName)),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.principal.type", "AWS"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.principal.identifiers.#", "1"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.principal.identifiers.0", rootUserArn),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.condition.StringLike.s3:prefix", bucketName),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.not_resource.#", "0"),
					resource.TestCheckResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.not_action.#", "0"),
					resource.TestCheckNoResourceAttr("storagegrid_bucket_policy.test", "policy.statement.0.not_principal"),
				),
			},
			// Delete testing is done automatically
		},
	})
}

func bucketPolicyConfiguration(bucketName, principalType string, principalIdentifiers []string, useCondition bool) string {
	var condition string
	if useCondition {
		condition = fmt.Sprintf(`
condition = {
	"StringLike" = {
		"s3:prefix" = "%s"
	}
}
`, bucketName)
	}

	bucketResource := fmt.Sprintf(`
resource "storagegrid_bucket" "test" {
	name = "%s"
}`, bucketName)

	var bucketPolicyResource string

	if len(principalIdentifiers) == 0 {
		bucketPolicyResource = fmt.Sprintf(`
resource "storagegrid_bucket_policy" "test" {
	bucket_name = storagegrid_bucket.test.name

	policy = {
		statement = [{
			sid = "test-sid"
			effect = "Allow"
			action = ["s3:ListBucket"]
			resource = ["arn:aws:s3:::%s", "arn:aws:s3:::%s/*"]
			principal = {
				type = "%s"
			}
			%s
		}]
	}
}
`, bucketName, bucketName, principalType, condition)
	} else {
		bucketPolicyResource = fmt.Sprintf(`
resource "storagegrid_bucket_policy" "test" {
	bucket_name = storagegrid_bucket.test.name

	policy = {
		statement = [{
			sid = "test-sid"
			effect = "Allow"
			action = ["s3:ListBucket"]
			resource = ["arn:aws:s3:::%s", "arn:aws:s3:::%s/*"]
			principal = {
				type = "%s"
				identifiers = [%s]
			}
			%s
		}]
	}
}
`, bucketName, bucketName, principalType, strings.Join(principalIdentifiers, ", "), condition)
	}

	return fmt.Sprintf("%s\n%s", bucketResource, bucketPolicyResource)
}
