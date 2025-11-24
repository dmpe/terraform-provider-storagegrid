// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	defaultRegion = "us-east-1"
)

func TestBucketResource_DefaultRegion(t *testing.T) {
	bucketName := fmt.Sprintf("tf-provider-acc-test-bucket-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"storagegrid": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Create
			{
				Config: testBucketResource(bucketName, nil, nil),
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
				Config: testBucketResource(bucketName, &region, nil),
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

func TestBucketResource_ObjectLockConfiguration(t *testing.T) {
	t.Run("compliance mode", func(t *testing.T) {
		bucketName := fmt.Sprintf("tf-provider-acc-test-bucket-%d", time.Now().Unix())

		objectLockConfiguration := ObjectLockConfiguration{
			Mode: types.StringValue("compliance"),
			Days: types.Int64Value(30),
		}
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				"storagegrid": providerserver.NewProtocol6WithError(New("test")()),
			},
			Steps: []resource.TestStep{
				// Create
				{
					Config: testBucketResource(bucketName, nil, &objectLockConfiguration),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "name", bucketName),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "region", defaultRegion),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "object_lock_configuration.mode", objectLockConfiguration.Mode.ValueString()),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "object_lock_configuration.days", strconv.FormatInt(objectLockConfiguration.Days.ValueInt64(), 10)),
						resource.TestCheckNoResourceAttr("storagegrid_bucket.test", "object_lock_configuration.years"),
					),
				},
				// Update retention days
				{
					Config: testBucketResource(bucketName, nil, &ObjectLockConfiguration{
						Mode: types.StringValue("compliance"),
						Days: types.Int64Value(15),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "name", bucketName),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "region", defaultRegion),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "object_lock_configuration.mode", objectLockConfiguration.Mode.ValueString()),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "object_lock_configuration.days", "15"),
						resource.TestCheckNoResourceAttr("storagegrid_bucket.test", "object_lock_configuration.years"),
					),
				},
				// Update mode
				{
					Config: testBucketResource(bucketName, nil, &ObjectLockConfiguration{
						Mode: types.StringValue("governance"),
						Days: types.Int64Value(15),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "name", bucketName),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "region", defaultRegion),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "object_lock_configuration.mode", "governance"),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "object_lock_configuration.days", "15"),
						resource.TestCheckNoResourceAttr("storagegrid_bucket.test", "object_lock_configuration.years"),
					),
				},
				// Update retention years
				{
					Config: testBucketResource(bucketName, nil, &ObjectLockConfiguration{
						Mode:  types.StringValue("governance"),
						Years: types.Int64Value(1),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "name", bucketName),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "region", defaultRegion),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "object_lock_configuration.mode", "governance"),
						resource.TestCheckNoResourceAttr("storagegrid_bucket.test", "object_lock_configuration.days"),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "object_lock_configuration.years", "1"),
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
	})

	t.Run("governance mode", func(t *testing.T) {
		bucketName := fmt.Sprintf("tf-provider-acc-test-bucket-%d", time.Now().Unix())

		objectLockConfiguration := ObjectLockConfiguration{
			Mode: types.StringValue("governance"),
			Days: types.Int64Value(30),
		}
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				"storagegrid": providerserver.NewProtocol6WithError(New("test")()),
			},
			Steps: []resource.TestStep{
				// Create
				{
					Config: testBucketResource(bucketName, nil, &objectLockConfiguration),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "name", bucketName),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "region", defaultRegion),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "object_lock_configuration.mode", objectLockConfiguration.Mode.ValueString()),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "object_lock_configuration.days", strconv.FormatInt(objectLockConfiguration.Days.ValueInt64(), 10)),
						resource.TestCheckNoResourceAttr("storagegrid_bucket.test", "object_lock_configuration.years"),
					),
				},
				// Update retention days
				{
					Config: testBucketResource(bucketName, nil, &ObjectLockConfiguration{
						Mode: types.StringValue("governance"),
						Days: types.Int64Value(15),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "name", bucketName),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "region", defaultRegion),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "object_lock_configuration.mode", objectLockConfiguration.Mode.ValueString()),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "object_lock_configuration.days", "15"),
						resource.TestCheckNoResourceAttr("storagegrid_bucket.test", "object_lock_configuration.years"),
					),
				},
				// Update mode
				{
					Config: testBucketResource(bucketName, nil, &ObjectLockConfiguration{
						Mode: types.StringValue("compliance"),
						Days: types.Int64Value(15),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "name", bucketName),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "region", defaultRegion),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "object_lock_configuration.mode", "compliance"),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "object_lock_configuration.days", "15"),
						resource.TestCheckNoResourceAttr("storagegrid_bucket.test", "object_lock_configuration.years"),
					),
				},
				// Update retention years
				{
					Config: testBucketResource(bucketName, nil, &ObjectLockConfiguration{
						Mode:  types.StringValue("compliance"),
						Years: types.Int64Value(1),
					}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "name", bucketName),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "region", defaultRegion),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "object_lock_configuration.mode", "compliance"),
						resource.TestCheckNoResourceAttr("storagegrid_bucket.test", "object_lock_configuration.days"),
						resource.TestCheckResourceAttr("storagegrid_bucket.test", "object_lock_configuration.years", "1"),
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
	})
}

func TestBucketResource_ObjectLockConfiguration_Versioning(t *testing.T) {
	bucketName := fmt.Sprintf("tf-provider-acc-test-bucket-%d", time.Now().Unix())

	objectLockConfiguration := ObjectLockConfiguration{
		Mode: types.StringValue("governance"),
		Days: types.Int64Value(30),
	}

	config := fmt.Sprintf(`
%s
resource "storagegrid_bucket_versioning" "test" {
	bucket_name = storagegrid_bucket.test.name
	status = "Enabled"
}
`, testBucketResource(bucketName, nil, &objectLockConfiguration))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"storagegrid": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Create
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("storagegrid_bucket.test", "name", bucketName),
					resource.TestCheckResourceAttr("storagegrid_bucket.test", "region", defaultRegion),
					resource.TestCheckResourceAttr("storagegrid_bucket.test", "object_lock_configuration.mode", objectLockConfiguration.Mode.ValueString()),
					resource.TestCheckResourceAttr("storagegrid_bucket.test", "object_lock_configuration.days", strconv.FormatInt(objectLockConfiguration.Days.ValueInt64(), 10)),
					resource.TestCheckNoResourceAttr("storagegrid_bucket.test", "object_lock_configuration.years"),
				),
			},
			// Delete testing is done automatically
		},
	})
}

func testBucketResource(name string, region *string, objectLockConfiguration *ObjectLockConfiguration) string {
	builder := strings.Builder{}
	builder.WriteString("resource \"storagegrid_bucket\" \"test\" {\n")
	builder.WriteString(fmt.Sprintf("\tname = \"%s\"", name))
	if region != nil {
		builder.WriteString(fmt.Sprintf("\n\tregion = \"%s\"", *region))
	}
	if objectLockConfiguration != nil {
		builder.WriteString(fmt.Sprintf("\n\tobject_lock_configuration {\n\t\tmode = \"%s\"", objectLockConfiguration.Mode.ValueString()))
		if objectLockConfiguration.Days.ValueInt64() != 0 {
			builder.WriteString(fmt.Sprintf("\n\t\tdays = %d", objectLockConfiguration.Days.ValueInt64()))
		}
		if objectLockConfiguration.Years.ValueInt64() != 0 {
			builder.WriteString(fmt.Sprintf("\n\t\tyears = %d", objectLockConfiguration.Years.ValueInt64()))
		}
		builder.WriteString("\n\t}")
	}
	builder.WriteString("\n}")
	return builder.String()
}
