// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
	client *BucketClient
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
		Blocks: map[string]schema.Block{
			"object_lock_configuration": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						Computed:    true,
						Description: "The object lock retention mode. Can be 'compliance' or 'governance'.",
					},
					"days": schema.Int64Attribute{
						Computed:    true,
						Description: "The number of days for which objects in the bucket are retained. Required if mode is 'compliance' or 'governance'.",
					},
					"years": schema.Int64Attribute{
						Computed:    true,
						Description: "The number of years for which objects in the bucket are retained. Required if mode is 'compliance' or 'governance'.",
					},
				},
				MarkdownDescription: "Object Lock configuration for the bucket. Will only be set if object locking is enabled for the bucket.",
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

	d.client = NewBucketClient(client)
}

func (d *bucketDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state BucketResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bucket, err := d.client.Read(ctx, state.Name.ValueString())
	if err != nil {
		if errors.Is(err, ErrBucketNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading StorageGrid container", err.Error())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &bucket)...)
}
