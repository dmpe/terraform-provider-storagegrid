// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &bucketVersioningResource{}
	_ resource.ResourceWithConfigure   = &bucketVersioningResource{}
	_ resource.ResourceWithImportState = &bucketVersioningResource{}
	_ resource.ResourceWithModifyPlan  = &bucketVersioningResource{}
)

// NewBucketVersioningResource returns a new resource instance.
func NewBucketVersioningResource() resource.Resource {
	return &bucketVersioningResource{}
}

// bucketResource defines the resource implementation.
type bucketVersioningResource struct {
	client *S3GridClient
}

func (r *bucketVersioningResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket_versioning"
}

func (r *bucketVersioningResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Manage object versioning for the named bucket.
If no versioning for bucket objects is required, do not create this resource, because disabled object versioning is the default,
and because the API produces an error if you attempt to create a resource with versioning status "Disabled".
`,
		Attributes: map[string]schema.Attribute{
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Required:    true,
				Description: "The status of versioning for the bucket. Can be 'Enabled', 'Suspended' or Disabled.",
				Validators: []validator.String{
					stringvalidator.OneOf("Enabled", "Suspended", "Disabled"),
				},
			},
		},
	}
}

func (r *bucketVersioningResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bucketVersioningResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BucketVersioningResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.upsert(plan)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
}

func (r *bucketVersioningResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BucketVersioningResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	read, err := r.read(ctx, state.BucketName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, read)...)
}

func (r *bucketVersioningResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BucketVersioningResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.upsert(plan)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
}

func (r *bucketVersioningResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	// NOOP
	// Applying this resource destruction will only remove the resource from the Terraform state without actually
	// disabling versioning on the bucket.
	// Consider suspending the versioning instead as it is not possible to disable versioning on a previously versioned
	// bucket.
}

func (r *bucketVersioningResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	state, err := r.read(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing StorageGrid container", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bucketVersioningResource) ModifyPlan(_ context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// check if the resource is planned for destruction
	if req.Plan.Raw.IsNull() {
		resp.Diagnostics.AddWarning(
			"Resource Destruction Considerations",
			"Applying this resource destruction will only remove the resource from the Terraform state without actually disabling versioning on the bucket. Consider suspending the versioning instead as it is not possible to disable versioning on a previously versioned bucket.",
		)
	}
}

func (r *bucketVersioningResource) read(ctx context.Context, bucketName string) (*BucketVersioningResourceModel, error) {
	tflog.Debug(ctx, "1. Get refreshed bucket information.")
	endpoint := fmt.Sprintf("%s/%s/versioning", api_buckets, bucketName)
	respBody, _, _, err := r.client.SendRequest("GET", endpoint, nil, 200)
	if err != nil {
		return nil, fmt.Errorf("unable to read bucket versioning: %w", err)
	}

	var returnBody BucketVersioningApiResponseModel
	tflog.Debug(ctx, "2. Unmarshal bucket information to JSON body.")
	if err := json.Unmarshal(respBody, &returnBody); err != nil {
		return nil, fmt.Errorf("unable to parse bucket versioning read response: %w", err)
	}

	return &BucketVersioningResourceModel{
		BucketName: types.StringValue(bucketName),
		Status:     types.StringValue(returnBody.Status()),
	}, nil
}

func (r *bucketVersioningResource) upsert(data BucketVersioningResourceModel) (*BucketVersioningResourceModel, error) {
	endpoint := fmt.Sprintf("%s/%s/versioning", api_buckets, data.BucketName.ValueString())
	httpResp, _, _, err := r.client.SendRequest("PUT", endpoint, data.ToBucketVersioningApiRequestModel(), 200)
	if err != nil {
		return nil, fmt.Errorf("unable to create or update bucket versioning: %w", err)
	}

	var returnBody BucketVersioningApiResponseModel
	if err := json.Unmarshal(httpResp, &returnBody); err != nil {
		return nil, fmt.Errorf("unable to unmarshal create or update bucket versioning response: %w", err)
	}

	return &BucketVersioningResourceModel{
		BucketName: data.BucketName,
		Status:     types.StringValue(returnBody.Status()),
	}, nil
}
