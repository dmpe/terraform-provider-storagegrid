// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &usersDataSource{}
var _ datasource.DataSourceWithConfigure = &usersDataSource{}

func NewUsersDataSource() datasource.DataSource {
	return &usersDataSource{}
}

// usersDataSource defines the data source implementation.
type usersDataSource struct {
	client *S3GridClient
}

func (d *usersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *usersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Fetch all users - a data source",

		Attributes: map[string]schema.Attribute{
			"data": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "the response data for the request (required on success and optional on error; type and content vary by request)",
				MarkdownDescription: "the response data for the request (required on success and optional on error; type and content vary by request)",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"unique_name": schema.StringAttribute{
							Computed: true,
						},
						fl_name: schema.StringAttribute{
							Computed: true,
						},
						"disable": schema.BoolAttribute{
							Computed: true,
						},
						"account_id": schema.StringAttribute{
							Computed: true,
						},
						"id": schema.StringAttribute{
							Computed: true,
						},
						"federated": schema.BoolAttribute{
							Computed: true,
						},
						"user_urn": schema.StringAttribute{
							Computed: true,
						},
						"member_of": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *usersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state usersDataSourceModel
	var jsonData UsersDataModel
	var newDiags diag.Diagnostics

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "1. Fetch all users from tenant.")
	rp, _, _, err := d.client.SendRequest("GET", api_users, nil, 200)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}

	// Map response body to model
	if err := json.Unmarshal(rp, &jsonData); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "2. Mapping data to TF state.")
	for _, item := range jsonData.Data {
		groupMembers := []types.String{}

		for _, s3Pol := range item.MemberOf {
			groupMembers = append(groupMembers, types.StringValue(s3Pol))
		}

		usersData := &usersDataSourceDataModel{
			UniqueName: types.StringValue(item.UniqueName),
			FullName:   types.StringValue(item.FullName),
			Disable:    types.BoolValue(item.Disable),
			AccountId:  types.StringValue(item.AccountId),
			ID:         types.StringValue(item.ID),
			Federated:  types.BoolValue(item.Federated),
			UserURN:    types.StringValue(item.UserURN),
			MemberOf:   groupMembers,
		}

		state.Data = append(state.Data, usersData)
	}

	resp.Diagnostics.Append(newDiags...)
	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read all users data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
