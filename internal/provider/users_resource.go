// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &usersResource{}
	_ resource.ResourceWithConfigure   = &usersResource{}
	_ resource.ResourceWithImportState = &usersResource{}
)

func NewUsersResource() resource.Resource {
	return &usersResource{}
}

// NewUsersResource defines the resource implementation.
type usersResource struct {
	client *S3GridClient
}

func (r *usersResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (r *usersResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		MarkdownDescription: "Users resource",
		Attributes: map[string]schema.Attribute{
			unique_name: schema.StringAttribute{
				Required: true,
			},
			"full_name": schema.StringAttribute{
				Required: true,
			},
			"disable": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"member_of": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"federated": schema.BoolAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"user_urn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			id: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *usersResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*S3GridClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *usersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan usersDataSourceDataModel
	var returnBody UsersDataModelSingle

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "1. Create to json body and fill it with the passed variables.")
	groupMembers := []string{}
	for _, member := range plan.MemberOf {
		groupMembers = append(groupMembers, member.ValueString())
	}
	body := &UserModelPostRequest{
		FullName:   plan.FullName.ValueString(),
		UniqueName: plan.UniqueName.ValueString(),
		Disable:    plan.Disable.ValueBool(),
		MemberOf:   groupMembers,
	}

	tflog.Debug(ctx, "2. Execute Request against REST api.")
	httpResp, _, _, err := r.client.SendRequest("POST", api_users, body, 201)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "3. User has been created and now we unmarshal it to json object.")
	if err := json.Unmarshal(httpResp, &returnBody); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "4. Mapping json body back to the state file.")
	returnBodyGroupMembers := []types.String{}

	plan.ID = types.StringValue(returnBody.Data.ID)
	plan.AccountId = types.StringValue(returnBody.Data.AccountId)
	plan.FullName = types.StringValue(returnBody.Data.FullName)
	plan.UniqueName = types.StringValue(returnBody.Data.UniqueName)
	plan.UserURN = types.StringValue(returnBody.Data.UserURN)
	plan.Disable = types.BoolValue(returnBody.Data.Disable)
	plan.Federated = types.BoolValue(returnBody.Data.Federated)
	for _, singleMember := range returnBody.Data.MemberOf {
		returnBodyGroupMembers = append(returnBodyGroupMembers, types.StringValue(singleMember))
	}
	plan.MemberOf = returnBodyGroupMembers

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a new user")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *usersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state usersDataSourceDataModel
	var returnBody UsersDataModelSingle
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "1. Get refreshed user information.")
	respBody, _, _, err := r.client.SendRequest("GET", api_users+"/"+state.ID.ValueString(), nil, 200)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading StorageGrid user",
			"Could not read StorageGrid user ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "2. Unmarshal user information to JSON body.")
	if err := json.Unmarshal(respBody, &returnBody); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "3. Overwrite fields with refreshed information.")
	groupMembers := []types.String{}
	for _, singleMember := range returnBody.Data.MemberOf {
		groupMembers = append(groupMembers, types.StringValue(singleMember))
	}

	usersData := &usersDataSourceDataModel{
		UniqueName: types.StringValue(returnBody.Data.UniqueName),
		FullName:   types.StringValue(returnBody.Data.FullName),
		Disable:    types.BoolValue(returnBody.Data.Disable),
		AccountId:  types.StringValue(returnBody.Data.AccountId),
		ID:         types.StringValue(returnBody.Data.ID),
		Federated:  types.BoolValue(returnBody.Data.Federated),
		UserURN:    types.StringValue(returnBody.Data.UserURN),
		MemberOf:   groupMembers,
	}

	state = *usersData

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *usersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state usersDataSourceDataModel
	var plan usersDataSourceDataModel
	var returnBody UsersDataModelSingle
	groupMembers := []string{}

	// Read Terraform plan + state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var userID = state.ID.ValueString()

	tflog.Debug(ctx, "1. Create updated user information.")
	for _, member := range plan.MemberOf {
		groupMembers = append(groupMembers, member.ValueString())
	}
	body := &UserModelPostRequest{
		FullName:   plan.FullName.ValueString(),
		UniqueName: plan.UniqueName.ValueString(),
		Disable:    plan.Disable.ValueBool(),
		MemberOf:   groupMembers,
	}

	tflog.Debug(ctx, "2. Execute Request against REST api.")
	_, _, _, err := r.client.SendRequest("PUT", api_users+"/"+userID, body, 200)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user information, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "3. Get refreshed user information.")
	respBody, _, _, err := r.client.SendRequest("GET", api_users+"/"+userID, nil, 200)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading StorageGrid user",
			"Could not read StorageGrid user ID "+userID+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "4. Unmarshal user information to JSON body.")
	if err := json.Unmarshal(respBody, &returnBody); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "5. Overwrite fields with refreshed information.")
	returnUserMemberOf := []types.String{}

	for _, singleMember := range returnBody.Data.MemberOf {
		returnUserMemberOf = append(returnUserMemberOf, types.StringValue(singleMember))
	}

	usersData := &usersDataSourceDataModel{
		UniqueName: types.StringValue(returnBody.Data.UniqueName),
		FullName:   types.StringValue(returnBody.Data.FullName),
		Disable:    types.BoolValue(returnBody.Data.Disable),
		AccountId:  types.StringValue(returnBody.Data.AccountId),
		ID:         types.StringValue(returnBody.Data.ID),
		Federated:  types.BoolValue(returnBody.Data.Federated),
		UserURN:    types.StringValue(returnBody.Data.UserURN),
		MemberOf:   returnUserMemberOf,
	}

	state = *usersData

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *usersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state usersDataSourceDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// in order for us to delete it, we first need to retrieve the same user and its ID
	_, _, _, err := r.client.SendRequest("DELETE", api_users+"/"+state.ID.ValueString(), nil, 204)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting StorageGrid user",
			"Could not delete user, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *usersResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
