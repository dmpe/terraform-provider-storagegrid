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

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

	payload, diags := plan.ToBucketModel()
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	httpResp, _, _, err := r.client.SendRequest("POST", api_buckets, payload, 201)
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

	if (plan.ObjectLockConfiguration != nil && state.ObjectLockConfiguration == nil) || (plan.ObjectLockConfiguration == nil && state.ObjectLockConfiguration != nil) {
		resp.Diagnostics.AddError("Error Updating StorageGrid container", "Object Lock Configuration cannot be changed once set.")
		return
	}

	payload, diags := plan.ToBucketModel()
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if payload.S3ObjectLock == nil {
		return
	}

	_, _, _, err := r.client.SendRequest("PUT", fmt.Sprintf("%s/%s/object-lock", api_buckets, state.Name.ValueString()), payload.S3ObjectLock, 200)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating StorageGrid container", fmt.Sprintf("Unable to update StorageGrid container, got error: %s", err.Error()))
	}

	updatedObjectLockConfiguration, err := r.readObjectLockConfiguration(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Updating StorageGrid container", fmt.Sprintf("Unable to read updated StorageGrid container, got error: %s", err.Error()))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &BucketResourceModel{Name: state.Name, Region: state.Region, ObjectLockConfiguration: updatedObjectLockConfiguration})...)
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

	if olc.Days.ValueInt64() == 0 && olc.Years.ValueInt64() == 0 {
		resp.Diagnostics.AddError(
			"Invalid Object Lock Configuration",
			"Object Lock Configuration must specify either days or years.",
		)
		return
	}
}

func (r *bucketResource) read(ctx context.Context, bucketName string) (*BucketResourceModel, error) {
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
		v, err := r.readRegion(ctx, bucketName)
		regCh <- regionResult{val: v, err: err}
	}()

	go func() {
		v, err := r.readObjectLockConfiguration(ctx, bucketName)
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

func (r *bucketResource) readRegion(ctx context.Context, bucketName string) (*string, error) {
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
		return nil, &GenericError{Summary: "Client Error", Details: "Unable to parse region response, got error: " + err.Error()}
	}

	return &returnBody.Data.Region, nil
}

func (r *bucketResource) readObjectLockConfiguration(ctx context.Context, bucketName string) (*ObjectLockConfiguration, error) {
	tflog.Debug(ctx, "1. Get refreshed bucket information.")
	endpoint := fmt.Sprintf("%s/%s/object-lock", api_buckets, bucketName)
	respBody, _, respCode, err := r.client.SendRequest("GET", endpoint, nil, 200)
	if err != nil {
		if respCode == http.StatusNotFound {
			return nil, ErrBucketNotFound
		}
	}

	type retentionSettings struct {
		Mode  string  `json:"mode"`
		Days  *string `json:"days,omitempty"`
		Years *string `json:"years,omitempty"`
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

	objectLockConfiguration := ObjectLockConfiguration{
		Mode: types.StringValue(returnBody.Data.RetentionSettings.Mode),
	}

	if strDays := returnBody.Data.RetentionSettings.Days; strDays != nil && *strDays != "" {
		days, err := strconv.Atoi(*strDays)
		if err != nil {
			return nil, &GenericError{Summary: "Client Error", Details: "Unable to parse object lock configuration's retention days, got error: " + err.Error()}
		}
		objectLockConfiguration.Days = types.Int64Value(int64(days))
	}

	if strYears := returnBody.Data.RetentionSettings.Years; strYears != nil && *strYears != "" {
		years, err := strconv.Atoi(*strYears)
		if err != nil {
			return nil, &GenericError{Summary: "Client Error", Details: "Unable to parse object lock configuration's retention years, got error: " + err.Error()}
		}
		objectLockConfiguration.Years = types.Int64Value(int64(years))
	}

	return &objectLockConfiguration, nil
}
