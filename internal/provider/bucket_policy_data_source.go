// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &bucketPolicyDataSource{}
	_ datasource.DataSourceWithConfigure = &bucketPolicyDataSource{}
)

// NewBucketPolicyDataSource returns a new resource instance.
func NewBucketPolicyDataSource() datasource.DataSource {
	return &bucketPolicyDataSource{}
}

// bucketPolicyDataSource defines the data source implementation.
type bucketPolicyDataSource struct {
	client *S3GridClient
}

func (d *bucketPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket_policy"
}

func (d *bucketPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Retrieve the access policy for a specific bucket, providing insights into who has access and what actions they can perform.
`,
		Attributes: map[string]schema.Attribute{
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket",
			},
			"policy": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The ID of the bucket policy (currently not used).",
					},
					"version": schema.StringAttribute{
						Computed:    true,
						Description: "The version of the bucket policy (currently not used).",
					},
					"statement": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"sid": schema.StringAttribute{
									Computed:    true,
									Description: "the unique identifier for the statement",
								},
								"effect": schema.StringAttribute{
									Computed:    true,
									Description: "the specific result of the statement (either an allow or an explicit deny)",
								},
								act: schema.ListAttribute{
									ElementType: types.StringType,
									Computed:    true,
									Description: "the specific actions that will be allowed",
								},
								n_act: schema.ListAttribute{
									ElementType: types.StringType,
									Computed:    true,
									Description: "the specific exceptional actions",
								},
								res: schema.ListAttribute{
									ElementType: types.StringType,
									Computed:    true,
									Description: "the objects that the statement covers",
								},
								n_res: schema.ListAttribute{
									ElementType: types.StringType,
									Computed:    true,
									Description: "the objects that the statement does not cover",
								},
								"condition": schema.MapAttribute{
									Computed:    true,
									Description: "the conditions that define when this policy statement will apply. (A Condition element can contain multiple conditions. Each condition consists of a Condition Type and a Condition Value.)",
									ElementType: types.MapType{}.WithElementType(types.StringType),
								},
								"principal": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"type": schema.StringAttribute{
											Computed:    true,
											Description: "the type of principal that is allowed access",
										},
										"identifiers": schema.ListAttribute{
											ElementType: types.StringType,
											Computed:    true,
											Description: "the identifiers of the principal that is allowed access",
										},
									},
									MarkdownDescription: "The principal(s) that are allowed access to the bucket.\n\n" +
										"~> Specify either `principal` or `not_principal`, but not both.\n\n" +
										"-> To have Terraform render JSON containing `\"Principal\": \"*\"`, use `type = \"*\"` and set the identifiers to `null` (or omit it entirely).\n" +
										"To have Terraform render JSON containing `\"Principal\": {\"AWS\": \"*\"}`, use `type = \"AWS\"` and `identifiers = [\"*\"]`.\n" +
										"If you want to specify a list of principals instead of a wildcard (`[\"*\"]`) specify a list of principal ARNs as identifiers.",
								},
								"not_principal": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"type": schema.StringAttribute{
											Computed:    true,
											Description: "the type of principal that is denied access",
										},
										"identifiers": schema.ListAttribute{
											ElementType: types.StringType,
											Computed:    true,
											Description: "the identifiers of the principal that is denied access",
										},
									},
									MarkdownDescription: "The principal(s) that are denied access to the bucket.\n\n" +
										"~> Specify either `principal` or `not_principal`, but not both.\n\n" +
										"-> To have Terraform render JSON containing `\"Principal\": \"\\*\"`, use `type = \"\\*\"` and set the identifiers to `null` (or omit it entirely).\n" +
										"To have Terraform render JSON containing `\"Principal\": {\"AWS\": \"\\*\"}`, use `type = \"AWS\"` and `identifiers = [\"\\*\"]`.\n" +
										"If you want to specify a list of principals instead of a wildcard (`[\"\\*\"]`) specify a list of principal ARNs as identifiers.",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *bucketPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*S3GridClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *bucketPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state BucketPolicyResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	read := state.read(d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &read)...)
}
