// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &bucketDataSource{}
var _ datasource.DataSourceWithConfigure = &bucketDataSource{}

// NewBucketDataSource returns a new resource instance.
func NewBucketDataSource() datasource.DataSource {
	return &bucketDataSource{}
}

// bucketDataSource defines the data source implementation.
type bucketDataSource struct {
	client *S3GridClient
}

func (d *bucketDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket"
}

func (d *bucketDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch a bucket by its name - a data source",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket",
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The region of the bucket, defaults to the StorageGRID's default region",
			},
		},
	}
}

func (d *bucketDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *bucketDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state BucketResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "1. Get refreshed user information.")
	endpoint := fmt.Sprintf("%s/%s/region", api_buckets, state.Name.ValueString())
	respBody, _, _, err := d.client.SendRequest("GET", endpoint, nil, 200)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read StorageGrid container, got error: %s", err))
		return
	}

	type regionDataModel struct {
		Region string `json:"region"`
	}

	type regionReadModel struct {
		Data regionDataModel `json:"data"`
	}

	var returnBody regionReadModel

	tflog.Debug(ctx, "2. Unmarshal bucket information to JSON body.")
	if err := json.Unmarshal(respBody, &returnBody); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "3. Overwrite fields with refreshed information.")
	bucket := BucketResourceModel{
		Name:   types.StringValue(state.Name.ValueString()),
		Region: types.StringValue(returnBody.Data.Region),
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &bucket)...)
}
