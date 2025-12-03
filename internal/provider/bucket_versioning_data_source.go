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
var _ datasource.DataSource = &bucketVersioningDataSource{}
var _ datasource.DataSourceWithConfigure = &bucketVersioningDataSource{}

// NewBucketVersioningDataSource returns a new resource instance.
func NewBucketVersioningDataSource() datasource.DataSource {
	return &bucketVersioningDataSource{}
}

// bucketVersioningDataSource defines the data source implementation.
type bucketVersioningDataSource struct {
	client *S3GridClient
}

func (d *bucketVersioningDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket_versioning"
}

func (d *bucketVersioningDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch the bucket configuration for object versioning of the named bucket.",
		Attributes: map[string]schema.Attribute{
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket to fetch versioning configuration for.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of versioning for the bucket. Is either 'Enabled', 'Suspended' or Disabled.",
			},
		},
	}
}

func (d *bucketVersioningDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *bucketVersioningDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state BucketVersioningResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "1. Get refreshed bucket information.")
	endpoint := fmt.Sprintf("%s/%s/versioning", api_buckets, state.BucketName.ValueString())
	respBody, _, _, err := d.client.SendRequest("GET", endpoint, nil, 200)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	var returnBody BucketVersioningApiResponseModel
	tflog.Debug(ctx, "2. Unmarshal bucket information to JSON body.")
	if err := json.Unmarshal(respBody, &returnBody); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse bucket versioning response, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "3. Overwrite fields with refreshed information.")
	bucket := BucketVersioningResourceModel{
		BucketName: types.StringValue(state.BucketName.ValueString()),
		Status:     types.StringValue(returnBody.Status()),
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &bucket)...)
}
