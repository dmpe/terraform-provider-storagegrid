// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &s3AccessSecretKeyCurrentUserResource{}
	_ resource.ResourceWithConfigure   = &s3AccessSecretKeyCurrentUserResource{}
	_ resource.ResourceWithImportState = &s3AccessSecretKeyCurrentUserResource{}
)

func NewS3AccessSecretKeyCurrentUserResource() resource.Resource {
	return &s3AccessSecretKeyCurrentUserResource{}
}

// News3AccessSecretKeyCurrentUserResource defines the resource implementation.
type s3AccessSecretKeyCurrentUserResource struct {
	client *S3GridClient
}

func (r *s3AccessSecretKeyCurrentUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_access_key_current_user"
}

func (r *s3AccessSecretKeyCurrentUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Create S3 access and secret key pair for current user - a resource",
		Attributes: map[string]schema.Attribute{
			"user_uuid": schema.StringAttribute{
				Required: true,
			},
			"expires": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
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
			"access_key": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secret_access_key": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *s3AccessSecretKeyCurrentUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *s3AccessSecretKeyCurrentUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan S3AccessKeyResourceModel
	var returnBody UserIDS3AccessSecretKeySingle

	var expiresConfig types.String

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("expires"), &expiresConfig)...)
	tflog.Debug(ctx, "1. Create to json body and fill it with the passed variables.")

	body := &UserIDS3AccessSecretKeysCreateJson{
		Expires: expiresConfig.ValueString(),
	}

	tflog.Debug(ctx, "2. Execute Request against REST api.")
	httpResp, _, _, err := r.client.SendRequest("POST", api_users+"/current-user"+api_s3_suffix, body, 201)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "3. S3 keys have been created and now we unmarshal it to json object.")
	if err := json.Unmarshal(httpResp, &returnBody); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}
	fmt.Println(returnBody.Data)
	tflog.Debug(ctx, "4. Mapping json body back to the state file.")
	acsKeyData := &S3AccessKeyResourceModel{
		ID:              types.StringValue(returnBody.Data.ID),
		AccountId:       types.StringValue(returnBody.Data.AccountId),
		DisplayName:     types.StringValue(returnBody.Data.DisplayName),
		UserURN:         types.StringValue(returnBody.Data.UserURN),
		UserUUID:        types.StringValue(returnBody.Data.UserUUID),
		Expires:         types.StringValue(returnBody.Data.Expires),
		AccessKey:       types.StringValue(returnBody.Data.AccessKey),
		SecretAccessKey: types.StringValue(returnBody.Data.SecretAccessKey),
	}
	plan = *acsKeyData

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a new s3 access keys")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *s3AccessSecretKeyCurrentUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state S3AccessKeyResourceModel
	var returnBody UserIDS3AccessSecretKeySingle

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "1. Get refreshed access key information.")
	respBody, _, _, err := r.client.SendRequest("GET", api_users+"/current-user"+api_s3_suffix+"/"+state.AccessKey.ValueString(), nil, 200)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading StorageGrid access key",
			"Could not read StorageGrid access key "+state.AccessKey.ValueString()+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "2. Unmarshal user information to JSON body.")
	if err := json.Unmarshal(respBody, &returnBody); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "3. Overwrite fields with refreshed information.")
	accessKeysReadOp := &S3AccessKeyResourceModel{
		ID:              types.StringValue(returnBody.Data.ID),
		AccountId:       types.StringValue(returnBody.Data.AccountId),
		DisplayName:     types.StringValue(returnBody.Data.DisplayName),
		UserURN:         types.StringValue(returnBody.Data.UserURN),
		UserUUID:        types.StringValue(returnBody.Data.UserUUID),
		Expires:         types.StringValue(returnBody.Data.Expires),
		AccessKey:       state.AccessKey,
		SecretAccessKey: state.SecretAccessKey,
	}

	state = *accessKeysReadOp

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *s3AccessSecretKeyCurrentUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Trace(ctx, "There is nothing to update - cannot be done")
	resp.Diagnostics.AddError("Not implemented", "There is nothing to update - cannot be done")
}

func (r *s3AccessSecretKeyCurrentUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserIDS3AccessSecretKeysModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// in order for us to delete it, we first need to retrieve user id and access key
	_, _, _, err := r.client.SendRequest("DELETE", api_users+"/current-user"+api_s3_suffix+"/"+state.Data.AccessKey.ValueString(), nil, 204)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting StorageGrid access key",
			"Could not delete access key, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *s3AccessSecretKeyCurrentUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("access_key"), req, resp)
}
