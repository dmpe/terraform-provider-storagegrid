// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                   = &bucketResource{}
	_ resource.ResourceWithConfigure      = &bucketResource{}
	_ resource.ResourceWithImportState    = &bucketResource{}
	_ resource.ResourceWithValidateConfig = &bucketResource{}
)

// NewBucketResource returns a new resource instance.
func NewBucketResource() resource.Resource {
	return &bucketResource{}
}

// bucketResource defines the resource implementation.
type bucketResource struct {
	client *BucketClient
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
		Blocks: map[string]schema.Block{
			"object_lock_configuration": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						Optional:    true,
						Description: "The object lock retention mode. Can be 'compliance' or 'governance'.",
						Validators: []validator.String{
							stringvalidator.OneOf("compliance", "governance"),
						},
					},
					"days": schema.Int64Attribute{
						Optional:    true,
						Description: "The number of days for which objects in the bucket are retained.",
						Validators: []validator.Int64{
							int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("years")),
						},
					},
					"years": schema.Int64Attribute{
						Optional:    true,
						Description: "The number of years for which objects in the bucket are retained.",
						Validators: []validator.Int64{
							int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("days")),
						},
					},
				},
				MarkdownDescription: `
Object Lock configuration for the bucket. Can only be set when creating a new bucket. If object locking is supposed to be disabled, omit the object lock configuration.

**Note:**
- Specify either 'days' or 'years', but not both if object locking is enabled.
- If object locking is enabled, object versioning will be enabled by default as well.
  It's safe to provide a "storagegrid_bucket_versioning" resource with status "Enabled" additionally.
`,
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

	r.client = NewBucketClient(client)
}

func (r *bucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BucketResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	read, err := r.client.Create(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating StorageGrid container", err.Error())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, read)...)
}

func (r *bucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state BucketResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bucket, err := r.client.Read(ctx, state.Name.ValueString())
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

func (r *bucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This is a noop for changes of the bucket's name and region as a resource re-creation is enforced.
	// Therefore, this update case is completely ignored, and we just deal with modifications of the object lock configuration.

	var plan BucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state BucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.Update(ctx, plan, state)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating StorageGrid container", fmt.Sprintf("Unable to update StorageGrid container, got error: %s", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
}

func (r *bucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete(state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting StorageGrid container",
			fmt.Sprintf("Could not delete bucket, unexpected error: %s", err.Error()),
		)
		return
	}
}

func (r *bucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	state, err := r.client.Read(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error Importing StorageGrid container", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ValidateConfig validates the configuration for the resource.
// It ensures either years or days are set for the `object_lock_configuration` block, but not both, and also not none.
func (r *bucketResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config BucketResourceModel
	if diags := req.Config.Get(ctx, &config); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if config.ObjectLockConfiguration == nil {
		return
	}

	olc := config.ObjectLockConfiguration

	if olc.Days.IsUnknown() || olc.Years.IsUnknown() {
		return
	}

	if olc.Days.ValueInt64() == 0 && olc.Years.ValueInt64() == 0 {
		resp.Diagnostics.AddError(
			"Invalid Object Lock Configuration",
			"Object Lock Configuration must specify either days or years.",
		)
		return
	}
}
