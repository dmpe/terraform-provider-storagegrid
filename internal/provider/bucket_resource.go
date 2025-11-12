// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &bucketResource{}
	_ resource.ResourceWithConfigure   = &bucketResource{}
	_ resource.ResourceWithImportState = &bucketResource{}
)

// NewBucketResource returns a new resource instance.
func NewBucketResource() resource.Resource {
	return &bucketResource{}
}

// bucketResource defines the resource implementation.
type bucketResource struct {
	client *S3GridClient
}

func (r *bucketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket"
}

func (r *bucketResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Create a new bucket - a resource",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				Description: "The region of the bucket, defaults to the StorageGRID's default region",
			},
		},
	}
}

func (r *bucketResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BucketResourceModel
	var returnBody BucketApiResponseModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, _, _, err := r.client.SendRequest("POST", api_buckets, plan.ToBucketModel(), 201)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create StorageGrid container, got error: %s", err.Error()))
		return
	}

	if err := json.Unmarshal(httpResp, &returnBody); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err.Error()))
		return
	}

	// if no region is provided, we need to read the resource to get the used default region.
	read, err := r.read(ctx, returnBody.Data.Name)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read newly created bucket, got error: %s", err.Error()))
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a new bucket")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &read)...)
}

func (r *bucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state BucketResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bucket, err := r.read(ctx, state.Name.ValueString())
	if err != nil {
		if errors.Is(err, ErrBucketNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading StorageGrid container", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &bucket)...)
}

func (r *bucketResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
	// noop
	// every change of either bucket name or region requires a new bucket to be created,
	// and the existing bucket is deleted
}

func (r *bucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, _, _, err := r.client.SendRequest("DELETE", api_buckets+"/"+state.Name.ValueString(), nil, 204); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting StorageGrid container",
			"Could not delete bucket, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *bucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	state, err := r.read(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing StorageGrid container", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bucketResource) read(ctx context.Context, bucketName string) (*BucketResourceModel, error) {
	tflog.Debug(ctx, "1. Get refreshed bucket information.")
	endpoint := fmt.Sprintf("%s/%s/region", api_buckets, bucketName)
	respBody, _, respCode, err := r.client.SendRequest("GET", endpoint, nil, 200)
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
		return nil, &GenericError{Summary: "Client Error", Details: "Unable to parse response, got error: " + err.Error()}
	}

	tflog.Debug(ctx, "3. Overwrite fields with refreshed information.")
	return &BucketResourceModel{
		Name:   types.StringValue(bucketName),
		Region: types.StringValue(returnBody.Data.Region),
	}, nil
}
