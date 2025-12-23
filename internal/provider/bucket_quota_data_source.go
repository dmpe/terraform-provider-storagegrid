// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &bucketQuotaDataSource{}
	_ datasource.DataSourceWithConfigure = &bucketQuotaDataSource{}
)

// NewBucketQuotaDataSource returns a new resource instance.
func NewBucketQuotaDataSource() datasource.DataSource {
	return &bucketQuotaDataSource{}
}

// bucketQuotaDataSource defines the data source implementation.
type bucketQuotaDataSource struct {
	client *S3GridClient
}

func (d *bucketQuotaDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket_quota"
}

func (d *bucketQuotaDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Fetch the bucket configuration for object quota of the named bucket.
If no quota is configured for the bucket, the 'object_bytes' attribute will be 'null'.
`,
		Attributes: map[string]schema.Attribute{
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket",
			},
			"object_bytes": schema.Int64Attribute{
				Computed:    true,
				Description: "The maximum number of bytes available for this bucket's objects. Represents a logical amount (object size), not a physical amount (size on disk).",
			},
		},
	}
}

func (d *bucketQuotaDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *bucketQuotaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state BucketQuotaResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	read, err := state.read(d.client)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading StorageGrid container object quota", err.Error())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &read)...)
}
