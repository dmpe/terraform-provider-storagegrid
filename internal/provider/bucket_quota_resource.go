// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &bucketQuotaResource{}
	_ resource.ResourceWithConfigure   = &bucketQuotaResource{}
	_ resource.ResourceWithImportState = &bucketQuotaResource{}
)

// NewBucketQuotaResource returns a new resource instance.
func NewBucketQuotaResource() resource.Resource {
	return &bucketQuotaResource{}
}

type bucketQuotaResource struct {
	client *S3GridClient
}

func (r *bucketQuotaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket_quota"
}

func (r *bucketQuotaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Define the maximum number of bytes available for this bucket's objects.
If no maximum number is required for the bucket's objects, do not specify this resource.
Similarly, removing an existing quota resource will set the maximum number of bytes to 'unlimited' for the referenced bucket.
`,
		Attributes: map[string]schema.Attribute{
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"object_bytes": schema.Int64Attribute{
				Required:    true,
				Description: "The maximum number of bytes available for this bucket's objects. Represents a logical amount (object size), not a physical amount (size on disk).",
				Validators: []validator.Int64{
					int64validator.Between(1, 1000000000000000000),
				},
			},
		},
	}
}

func (r *bucketQuotaResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bucketQuotaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BucketQuotaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	quota, err := plan.upsert(r.client)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, quota)...)
}

func (r *bucketQuotaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BucketQuotaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	quota, err := plan.upsert(r.client)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, quota)...)
}

func (r *bucketQuotaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BucketQuotaResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	read, err := state.read(r.client)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading StorageGrid container object quota", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, read)...)
}

func (r *bucketQuotaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BucketQuotaResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := state.delete(r.client)
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting StorageGrid container object quota", err.Error())
		return
	}
}

func (r *bucketQuotaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	model := BucketQuotaResourceModel{
		BucketName: types.StringValue(req.ID),
	}

	state, err := model.read(r.client)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing StorageGrid container object quota", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
