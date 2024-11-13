// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &userDataSource{}
var _ datasource.DataSourceWithConfigure = &userDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

// userDataSource defines the data source implementation.
type userDataSource struct {
	client *S3GridClient
}

func (d *userDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "user data source",
		Attributes: map[string]schema.Attribute{
			unique_name: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(
						path.MatchRelative().AtParent().AtName(id),
						path.MatchRelative().AtParent().AtName(unique_name),
					),
				},
			},
			"full_name": schema.StringAttribute{
				Computed: true,
			},
			"disable": schema.BoolAttribute{
				Computed: true,
			},
			"account_id": schema.StringAttribute{
				Computed: true,
			},
			id: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(
						path.MatchRelative().AtParent().AtName(id),
						path.MatchRelative().AtParent().AtName(unique_name),
					),
				},
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
	}
}

func (d *userDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state usersDataSourceDataModel
	var jsonData UsersDataModelSingle
	var newDiags diag.Diagnostics
	var idType types.String
	var uniqueNameType types.String
	var fullPath string
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("id"), &idType)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("unique_name"), &uniqueNameType)...)

	tflog.Debug(ctx, "1. Fetching user by id or by name.")
	if !state.ID.IsNull() {
		fullPath = api_users + "/" + idType.ValueString()
	} else {
		fullPath = api_users + "/" + uniqueNameType.ValueString()
	}
	rp, _, _, err := d.client.SendRequest("GET", fullPath, nil, 200)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "2. Map response body to model.")
	if err := json.Unmarshal(rp, &jsonData); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "3. Mapping data to TF state.")
	groupMembers := []types.String{}

	for _, singleMember := range jsonData.Data.MemberOf {
		groupMembers = append(groupMembers, types.StringValue(singleMember))
	}

	usersData := &usersDataSourceDataModel{
		UniqueName: types.StringValue(jsonData.Data.UniqueName),
		FullName:   types.StringValue(jsonData.Data.FullName),
		Disable:    types.BoolValue(jsonData.Data.Disable),
		AccountId:  types.StringValue(jsonData.Data.AccountId),
		ID:         types.StringValue(jsonData.Data.ID),
		Federated:  types.BoolValue(jsonData.Data.Federated),
		UserURN:    types.StringValue(jsonData.Data.UserURN),
		MemberOf:   groupMembers,
	}

	state = *usersData

	resp.Diagnostics.Append(newDiags...)
	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read users data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
