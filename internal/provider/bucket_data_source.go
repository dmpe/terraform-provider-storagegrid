// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

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

	d.client = client
}

func (d *bucketDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state BucketResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bucket, err := d.read(ctx, state.Name.ValueString())
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

func (d *bucketDataSource) read(ctx context.Context, bucketName string) (*BucketResourceModel, error) {
	// Read region and object lock configuration in parallel.
	type regionResult struct {
		val *string
		err error
	}
	type olcResult struct {
		val *ObjectLockConfiguration
		err error
	}

	regCh := make(chan regionResult, 1)
	olcCh := make(chan olcResult, 1)

	go func() {
		v, err := d.readRegion(ctx, bucketName)
		regCh <- regionResult{val: v, err: err}
	}()

	go func() {
		v, err := d.readObjectLockConfiguration(ctx, bucketName)
		olcCh <- olcResult{val: v, err: err}
	}()

	reg := <-regCh
	olc := <-olcCh

	combinedErrs := errors.Join(reg.err, olc.err)
	if combinedErrs != nil {
		return nil, combinedErrs
	}

	return &BucketResourceModel{
		Name:                    types.StringValue(bucketName),
		Region:                  types.StringValue(*reg.val),
		ObjectLockConfiguration: olc.val,
	}, nil
}

func (d *bucketDataSource) readRegion(ctx context.Context, bucketName string) (*string, error) {
	tflog.Debug(ctx, "1. Get refreshed bucket information.")
	endpoint := fmt.Sprintf("%s/%s/region", api_buckets, bucketName)
	respBody, _, respCode, err := d.client.SendRequest("GET", endpoint, nil, 200)
	if err != nil {
		if respCode == http.StatusNotFound {
			return nil, ErrBucketNotFound
		}
		return nil, &GenericError{Summary: "Error Reading StorageGrid container", Details: "Could not read StorageGrid container name " + bucketName + ": " + err.Error()}
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
		return nil, &GenericError{Summary: "Client Error", Details: "Unable to parse region response, got error: " + err.Error()}
	}

	return &returnBody.Data.Region, nil
}

func (d *bucketDataSource) readObjectLockConfiguration(ctx context.Context, bucketName string) (*ObjectLockConfiguration, error) {
	tflog.Debug(ctx, "1. Get refreshed bucket information.")
	endpoint := fmt.Sprintf("%s/%s/object-lock", api_buckets, bucketName)
	respBody, _, respCode, err := d.client.SendRequest("GET", endpoint, nil, 200)
	if err != nil {
		if respCode == http.StatusNotFound {
			return nil, ErrBucketNotFound
		}
	}

	type retentionSettings struct {
		Mode  string `json:"mode"`
		Days  string `json:"days,omitempty"`
		Years string `json:"years,omitempty"`
	}

	type s3ObjectLock struct {
		Enabled           bool              `json:"enabled"`
		RetentionSettings retentionSettings `json:"defaultRetentionSetting"`
	}

	type objectLockConfigurationReadModel struct {
		Data s3ObjectLock `json:"data"`
	}

	var returnBody objectLockConfigurationReadModel

	tflog.Debug(ctx, "2. Unmarshal bucket information to JSON body.")
	if err := json.Unmarshal(respBody, &returnBody); err != nil {
		return nil, &GenericError{Summary: "Client Error", Details: "Unable to parse object lock configuration response, got error: " + err.Error()}
	}

	if !returnBody.Data.Enabled {
		return nil, nil
	}

	var days, years int

	if strDays := returnBody.Data.RetentionSettings.Days; strDays != "" {
		days, err = strconv.Atoi(returnBody.Data.RetentionSettings.Days)
		if err != nil {
			return nil, &GenericError{Summary: "Client Error", Details: "Unable to parse object lock configuration's retention days, got error: " + err.Error()}
		}
	}

	if strYears := returnBody.Data.RetentionSettings.Years; strYears != "" {
		years, err = strconv.Atoi(returnBody.Data.RetentionSettings.Years)
		if err != nil {
			return nil, &GenericError{Summary: "Client Error", Details: "Unable to parse object lock configuration's retention years, got error: " + err.Error()}
		}
	}

	return &ObjectLockConfiguration{
		Mode:  types.StringValue(returnBody.Data.RetentionSettings.Mode),
		Days:  types.Int64Value(int64(days)),
		Years: types.Int64Value(int64(years)),
	}, nil
}
