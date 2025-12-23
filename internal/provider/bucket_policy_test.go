// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

var conditionOperators = map[string]attr.Value{
	"StringLike": types.MapValueMust(types.StringType, map[string]attr.Value{
		"s3:prefix": types.StringValue("test-bucket"),
	}),
	"StringEquals": types.MapValueMust(types.StringType, map[string]attr.Value{
		"s3:ExistingObjectTag/Name": types.StringValue("test-tag"),
		"aws:username":              types.StringValue("test-user"),
	}),
}
var condition = types.MapValueMust(types.MapType{}.WithElementType(types.StringType), conditionOperators)

func TestStatementResourceModel_JSON(t *testing.T) {
	tests := []struct {
		name     string
		model    BucketPolicyResourceModel
		expected string
	}{
		{
			name: "all fields with wildcard principal and one non principal AWS",
			model: BucketPolicyResourceModel{
				BucketName: types.StringValue("test-bucket"),
				Policy: &PolicyResourceModel{
					Id:      types.StringValue("test-id"),
					Version: types.StringValue("test-version"),
					Statement: []StatementResourceModel{
						{
							Sid:         types.StringValue("test-sid"),
							Effect:      types.StringValue("test-effect"),
							Action:      []types.String{types.StringValue("test-action")},
							NotAction:   []types.String{types.StringValue("test-not-action")},
							Resource:    []types.String{types.StringValue("test-resource")},
							NotResource: []types.String{types.StringValue("test-not-resource")},
							Condition:   condition,
							Principal: &PrincipalResourceModel{
								Type:        types.StringValue("*"),
								Identifiers: nil,
							},
							NotPrincipal: &PrincipalResourceModel{
								Type:        types.StringValue("AWS"),
								Identifiers: []types.String{types.StringValue("arn:aws:iam::123456789012:user/test-user")},
							},
						},
					},
				},
			},
			expected: `
{
	"policy": {
		"Id": "test-id",
		"Version": "test-version",
		"Statement": [{
			"Sid": "test-sid",
			"Effect": "test-effect",
			"Action": ["test-action"],
			"NotAction": ["test-not-action"],
			"Resource": ["test-resource"],
			"NotResource": ["test-not-resource"],
			"Condition": {
				"StringLike": {
					"s3:prefix": "test-bucket"
				},
				"StringEquals": {
					"s3:ExistingObjectTag/Name": "test-tag",
					"aws:username": "test-user"
				}
			},
			"Principal": "*",
			"NotPrincipal": {
				"AWS": "arn:aws:iam::123456789012:user/test-user"
			}
		}]
	}
}
`,
		},
		{
			name: "all fields with wildcard principal and wildcard non principal",
			model: BucketPolicyResourceModel{
				BucketName: types.StringValue("test-bucket"),
				Policy: &PolicyResourceModel{
					Id:      types.StringValue("test-id"),
					Version: types.StringValue("test-version"),
					Statement: []StatementResourceModel{
						{
							Sid:         types.StringValue("test-sid"),
							Effect:      types.StringValue("test-effect"),
							Action:      []types.String{types.StringValue("test-action")},
							NotAction:   []types.String{types.StringValue("test-not-action")},
							Resource:    []types.String{types.StringValue("test-resource")},
							NotResource: []types.String{types.StringValue("test-not-resource")},
							Condition:   condition,
							Principal: &PrincipalResourceModel{
								Type:        types.StringValue("*"),
								Identifiers: nil,
							},
							NotPrincipal: &PrincipalResourceModel{
								Type:        types.StringValue("AWS"),
								Identifiers: nil,
							},
						},
					},
				},
			},
			expected: `
{
	"policy": {
		"Id": "test-id",
		"Version": "test-version",
		"Statement": [{
			"Sid": "test-sid",
			"Effect": "test-effect",
			"Action": ["test-action"],
			"NotAction": ["test-not-action"],
			"Resource": ["test-resource"],
			"NotResource": ["test-not-resource"],
			"Condition": {
				"StringLike": {
					"s3:prefix": "test-bucket"
				},
				"StringEquals": {
					"s3:ExistingObjectTag/Name": "test-tag",
					"aws:username": "test-user"
				}
			},
			"Principal": "*",
			"NotPrincipal": {
				"AWS": "*"
			}
		}]
	}
}
`,
		},
		{
			name: "all fields with wildcard principal and two non principal AWS",
			model: BucketPolicyResourceModel{
				BucketName: types.StringValue("test-bucket"),
				Policy: &PolicyResourceModel{
					Id:      types.StringValue("test-id"),
					Version: types.StringValue("test-version"),
					Statement: []StatementResourceModel{
						{
							Sid:         types.StringValue("test-sid"),
							Effect:      types.StringValue("test-effect"),
							Action:      []types.String{types.StringValue("test-action")},
							NotAction:   []types.String{types.StringValue("test-not-action")},
							Resource:    []types.String{types.StringValue("test-resource")},
							NotResource: []types.String{types.StringValue("test-not-resource")},
							Condition:   condition,
							Principal: &PrincipalResourceModel{
								Type:        types.StringValue("*"),
								Identifiers: nil,
							},
							NotPrincipal: &PrincipalResourceModel{
								Type: types.StringValue("AWS"),
								Identifiers: []types.String{
									types.StringValue("arn:aws:iam::123456789012:user/test-user"),
									types.StringValue("arn:aws:iam::123456789012:user/test-user2"),
								},
							},
						},
					},
				},
			},
			expected: `
{
	"policy": {
		"Id": "test-id",
		"Version": "test-version",
		"Statement": [{
			"Sid": "test-sid",
			"Effect": "test-effect",
			"Action": ["test-action"],
			"NotAction": ["test-not-action"],
			"Resource": ["test-resource"],
			"NotResource": ["test-not-resource"],
			"Condition": {
				"StringLike": {
					"s3:prefix": "test-bucket"
				},
				"StringEquals": {
					"s3:ExistingObjectTag/Name": "test-tag",
					"aws:username": "test-user"
				}
			},
			"Principal": "*",
			"NotPrincipal": {
				"AWS": [
					"arn:aws:iam::123456789012:user/test-user",
					"arn:aws:iam::123456789012:user/test-user2"
				]
			}
		}]
	}
}
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var diags diag.Diagnostics

			buf := new(bytes.Buffer)
			enc := json.NewEncoder(buf)
			err := enc.Encode(test.model.toBucketPolicyApiModel(context.Background(), &diags))
			if err != nil {
				t.FailNow()
			}

			assert.False(t, diags.HasError())
			assert.JSONEq(t, test.expected, buf.String())
		})
	}
}

func TestNewBucketPolicyResourceModel(t *testing.T) {
	tests := []struct {
		name string
		json string
		want BucketPolicyResourceModel
	}{
		{
			name: "all fields with wildcard principal and one non principal AWS",
			json: `
{
	"data": {
		"policy": {
			"Id": "test-id",
			"Version": "test-version",
			"Statement": [{
				"Sid": "test-sid",
				"Effect": "test-effect",
				"Action": ["test-action"],
				"NotAction": ["test-not-action"],
				"Resource": ["test-resource"],
				"NotResource": ["test-not-resource"],
				"Condition": {
					"StringLike": {
						"s3:prefix": "test-bucket"
					},
					"StringEquals": {
						"s3:ExistingObjectTag/Name": "test-tag",
						"aws:username": "test-user"
					}
				},
				"Principal": "*",
				"NotPrincipal": {
					"AWS": "arn:aws:iam::123456789012:user/test-user"
				}
			}]
		}
	}
}
`,
			want: BucketPolicyResourceModel{
				BucketName: types.StringValue("test-bucket"),
				Policy: &PolicyResourceModel{
					Id:      types.StringValue("test-id"),
					Version: types.StringValue("test-version"),
					Statement: []StatementResourceModel{
						{
							Sid:         types.StringValue("test-sid"),
							Effect:      types.StringValue("test-effect"),
							Action:      []types.String{types.StringValue("test-action")},
							NotAction:   []types.String{types.StringValue("test-not-action")},
							Resource:    []types.String{types.StringValue("test-resource")},
							NotResource: []types.String{types.StringValue("test-not-resource")},
							Condition:   condition,
							Principal: &PrincipalResourceModel{
								Type:        types.StringValue("*"),
								Identifiers: nil,
							},
							NotPrincipal: &PrincipalResourceModel{
								Type:        types.StringValue("AWS"),
								Identifiers: []types.String{types.StringValue("arn:aws:iam::123456789012:user/test-user")},
							},
						},
					},
				},
			},
		},
		{
			name: "all fields with wildcard principal and wildcard non principal",
			json: `
{
	"data": {
		"policy": {
			"Id": "test-id",
			"Version": "test-version",
			"Statement": [{
				"Sid": "test-sid",
				"Effect": "test-effect",
				"Action": ["test-action"],
				"NotAction": ["test-not-action"],
				"Resource": ["test-resource"],
				"NotResource": ["test-not-resource"],
				"Condition": {
					"StringLike": {
						"s3:prefix": "test-bucket"
					},
					"StringEquals": {
						"s3:ExistingObjectTag/Name": "test-tag",
						"aws:username": "test-user"
					}
				},
				"Principal": "*",
				"NotPrincipal": {
					"AWS": "*"
				}
			}]
		}
	}
}
`,
			want: BucketPolicyResourceModel{
				BucketName: types.StringValue("test-bucket"),
				Policy: &PolicyResourceModel{
					Id:      types.StringValue("test-id"),
					Version: types.StringValue("test-version"),
					Statement: []StatementResourceModel{
						{
							Sid:         types.StringValue("test-sid"),
							Effect:      types.StringValue("test-effect"),
							Action:      []types.String{types.StringValue("test-action")},
							NotAction:   []types.String{types.StringValue("test-not-action")},
							Resource:    []types.String{types.StringValue("test-resource")},
							NotResource: []types.String{types.StringValue("test-not-resource")},
							Condition:   condition,
							Principal: &PrincipalResourceModel{
								Type:        types.StringValue("*"),
								Identifiers: nil,
							},
							NotPrincipal: &PrincipalResourceModel{
								Type:        types.StringValue("AWS"),
								Identifiers: nil,
							},
						},
					},
				},
			},
		},
		{
			name: "all fields with wildcard principal and two non principal AWS",
			json: `
{
	"data": {
		"policy": {
			"Id": "test-id",
			"Version": "test-version",
			"Statement": [{
				"Sid": "test-sid",
				"Effect": "test-effect",
				"Action": ["test-action"],
				"NotAction": ["test-not-action"],
				"Resource": ["test-resource"],
				"NotResource": ["test-not-resource"],
				"Condition": {
					"StringLike": {
						"s3:prefix": "test-bucket"
					},
					"StringEquals": {
						"s3:ExistingObjectTag/Name": "test-tag",
						"aws:username": "test-user"
					}
				},
				"Principal": "*",
				"NotPrincipal": {
					"AWS": [
						"arn:aws:iam::123456789012:user/test-user",
						"arn:aws:iam::123456789012:user/test-user2"
					]
				}
			}]
		}
	}
}
`,
			want: BucketPolicyResourceModel{
				BucketName: types.StringValue("test-bucket"),
				Policy: &PolicyResourceModel{
					Id:      types.StringValue("test-id"),
					Version: types.StringValue("test-version"),
					Statement: []StatementResourceModel{
						{
							Sid:         types.StringValue("test-sid"),
							Effect:      types.StringValue("test-effect"),
							Action:      []types.String{types.StringValue("test-action")},
							NotAction:   []types.String{types.StringValue("test-not-action")},
							Resource:    []types.String{types.StringValue("test-resource")},
							NotResource: []types.String{types.StringValue("test-not-resource")},
							Condition:   condition,
							Principal: &PrincipalResourceModel{
								Type:        types.StringValue("*"),
								Identifiers: nil,
							},
							NotPrincipal: &PrincipalResourceModel{
								Type: types.StringValue("AWS"),
								Identifiers: []types.String{
									types.StringValue("arn:aws:iam::123456789012:user/test-user"),
									types.StringValue("arn:aws:iam::123456789012:user/test-user2"),
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var diags diag.Diagnostics
			model := NewBucketPolicyResourceModel("test-bucket", []byte(test.json), &diags)
			assert.False(t, diags.HasError())
			assert.NotNil(t, model)
			assert.Equal(t, test.want, *model)
		})
	}
}
