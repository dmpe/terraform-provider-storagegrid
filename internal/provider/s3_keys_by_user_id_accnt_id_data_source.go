// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &s3UserIDSpecificAccessKeyDataSource{}
var _ datasource.DataSourceWithConfigure = &s3UserIDSpecificAccessKeyDataSource{}

func NewS3DataSource_ByUserID_AccountID() datasource.DataSource {
	return &s3UserIDSpecificAccessKeyDataSource{}
}

// s3UserIDSpecificAccessKeyDataSource defines the data source implementation.
type s3UserIDSpecificAccessKeyDataSource struct {
	client *S3GridClient
}

func (d *s3UserIDSpecificAccessKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_user_id_access_key"
}

func (d *s3UserIDSpecificAccessKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Access specific S3 access key for a user - a data source",
		Attributes: map[string]schema.Attribute{
			"user_uuid": schema.StringAttribute{
				Required: true,
			},
			"access_key": schema.StringAttribute{
				Required: true,
			},
			"data": schema.SingleNestedAttribute{
				Computed:            true,
				Description:         "the response data for the request (required on success and optional on error; type and content vary by request)",
				MarkdownDescription: "the response data for the request (required on success and optional on error; type and content vary by request)",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
					},
					"account_id": schema.StringAttribute{
						Computed: true,
					},
					"display_name": schema.StringAttribute{
						Computed: true,
					},
					"user_urn": schema.StringAttribute{
						Computed: true,
					},
					"user_uuid": schema.StringAttribute{
						Computed: true,
					},
					"expires": schema.StringAttribute{
						Computed: true,
					},
				},
			},
		},
	}
}

func (d *s3UserIDSpecificAccessKeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *s3UserIDSpecificAccessKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state UserIDS3AccessKeysModel
	var returnBody UserIDS3AccessKeySingle

	var newDiags diag.Diagnostics
	var userId types.String
	var accessKey types.String

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("user_uuid"), &userId)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("access_key"), &accessKey)...)

	tflog.Debug(ctx, "1. Fetch S3 access key by user id and access key.")
	rp, _, _, err := d.client.SendRequest("GET", api_users+"/"+userId.ValueString()+api_s3_suffix+"/"+accessKey.ValueString(), nil, 200)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "2. Map response body to model.")
	if err := json.Unmarshal(rp, &returnBody); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "3. Mapping data to TF state.")
	accessKeyData := &S3AccessKeyModel{
		ID:          types.StringValue(returnBody.Data.ID),
		AccountId:   types.StringValue(returnBody.Data.AccountId),
		DisplayName: types.StringValue(returnBody.Data.DisplayName),
		UserURN:     types.StringValue(returnBody.Data.UserURN),
		UserUUID:    types.StringValue(returnBody.Data.UserUUID),
		Expires:     types.StringValue(returnBody.Data.Expires),
	}
	userS3AccessKeyData := &UserIDS3AccessKeysModel{
		AccessKey: accessKey,
		UserUUID:  userId,
		Data:      accessKeyData,
	}
	state = *userS3AccessKeyData

	resp.Diagnostics.Append(newDiags...)
	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read user's access keys - data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
