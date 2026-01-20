// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &bucketPolicyResource{}
	_ resource.ResourceWithConfigure   = &bucketPolicyResource{}
	_ resource.ResourceWithImportState = &bucketPolicyResource{}
)

var emptyStringListValue basetypes.ListValue

func init() {
	eslv, err := basetypes.NewListValue(types.StringType, []attr.Value{})
	if err != nil {
		panic(err)
	}
	emptyStringListValue = eslv
}

// NewBucketPolicyResource returns a new resource instance.
func NewBucketPolicyResource() resource.Resource {
	return &bucketPolicyResource{}
}

type bucketPolicyResource struct {
	client *S3GridClient
}

func (r *bucketPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket_policy"
}

func (r *bucketPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Define access policies for the bucket, allowing fine-grained control over who can access and modify its contents.
`,
		Attributes: map[string]schema.Attribute{
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The ID of the bucket policy (currently not used).",
					},
					"version": schema.StringAttribute{
						Computed:    true,
						Description: "The version of the bucket policy (currently not used).",
					},
					"statement": schema.ListNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"sid": schema.StringAttribute{
									Optional:    true,
									Description: "the unique identifier for the statement",
								},
								"effect": schema.StringAttribute{
									Required:    true,
									Description: "the specific result of the statement (either an allow or an explicit deny)",
									Validators: []validator.String{
										stringvalidator.OneOf(
											"Allow",
											"Deny",
										),
									},
								},
								act: schema.ListAttribute{
									ElementType: types.StringType,
									Optional:    true,
									Description: "The specific actions that will be allowed. A statement must have either Action or NotAction.",
									Validators: []validator.List{
										listvalidator.AtLeastOneOf(
											path.MatchRelative().AtParent().AtName(act),
											path.MatchRelative().AtParent().AtName(n_act),
										),
									},
								},
								n_act: schema.ListAttribute{
									ElementType: types.StringType,
									Optional:    true,
									Computed:    true,
									Description: "the specific exceptional actions. A statement must have either Action or NotAction.",
									Validators: []validator.List{
										listvalidator.AtLeastOneOf(
											path.MatchRelative().AtParent().AtName(act),
											path.MatchRelative().AtParent().AtName(n_act),
										),
									},
									Default: listdefault.StaticValue(emptyStringListValue),
								},
								res: schema.ListAttribute{
									ElementType: types.StringType,
									Optional:    true,
									Description: "the objects that the statement covers. A statement must have either Resource or NotResource.",
									Validators: []validator.List{
										listvalidator.AtLeastOneOf(
											path.MatchRelative().AtParent().AtName(n_res),
											path.MatchRelative().AtParent().AtName(res),
										),
									},
								},
								n_res: schema.ListAttribute{
									ElementType: types.StringType,
									Optional:    true,
									Computed:    true,
									Description: "the objects that the statement does not cover. A statement must have either Resource or NotResource.",
									Validators: []validator.List{
										listvalidator.AtLeastOneOf(
											path.MatchRelative().AtParent().AtName(n_res),
											path.MatchRelative().AtParent().AtName(res),
										),
									},
									Default: listdefault.StaticValue(emptyStringListValue),
								},
								"condition": schema.MapAttribute{
									Optional:    true,
									Description: "the conditions that define when this policy statement will apply. (A Condition element can contain multiple conditions. Each condition consists of a Condition Type and a Condition Value.)",
									ElementType: types.MapType{}.WithElementType(types.StringType),
								},
								"principal": schema.SingleNestedAttribute{
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"type": schema.StringAttribute{
											Required:    true,
											Description: "the type of principal that is allowed access",
											Validators: []validator.String{
												stringvalidator.OneOf(
													"AWS", "*",
												),
											},
										},
										"identifiers": schema.ListAttribute{
											ElementType: types.StringType,
											Optional:    true,
											Description: "the identifiers of the principal that is allowed access",
										},
									},
									MarkdownDescription: "The principal(s) that are allowed access to the bucket.\n\n" +
										"~> Specify either `principal` or `not_principal`, but not both.\n\n" +
										"-> To have Terraform render JSON containing `\"Principal\": \"*\"`, use `type = \"*\"` and set the identifiers to `null` (or omit it entirely).\n" +
										"To have Terraform render JSON containing `\"Principal\": {\"AWS\": \"*\"}`, use `type = \"AWS\"` and `identifiers = [\"*\"]`.\n" +
										"If you want to specify a list of principals instead of a wildcard (`[\"*\"]`) specify a list of principal ARNs as identifiers.",
									Validators: []validator.Object{
										objectvalidator.AtLeastOneOf(
											path.MatchRelative().AtParent().AtName("principal"),
											path.MatchRelative().AtParent().AtName("not_principal"),
										),
									},
								},
								"not_principal": schema.SingleNestedAttribute{
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"type": schema.StringAttribute{
											Required:    true,
											Description: "the type of principal that is denied access",
											Validators: []validator.String{
												stringvalidator.OneOf(
													"AWS", "*",
												),
											},
										},
										"identifiers": schema.ListAttribute{
											ElementType: types.StringType,
											Optional:    true,
											Description: "the identifiers of the principal that is denied access",
										},
									},
									MarkdownDescription: "The principal(s) that are denied access to the bucket.\n\n" +
										"~> Specify either `principal` or `not_principal`, but not both.\n\n" +
										"-> To have Terraform render JSON containing `\"Principal\": \"*\"`, use `type = \"*\"` and set the identifiers to `null` (or omit it entirely).\n" +
										"To have Terraform render JSON containing `\"Principal\": {\"AWS\": \"*\"}`, use `type = \"AWS\"` and `identifiers = [\"*\"]`.\n" +
										"If you want to specify a list of principals instead of a wildcard (`[\"*\"]`) specify a list of principal ARNs as identifiers.",
									Validators: []validator.Object{
										objectvalidator.AtLeastOneOf(
											path.MatchRelative().AtParent().AtName("principal"),
											path.MatchRelative().AtParent().AtName("not_principal"),
										),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *bucketPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bucketPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BucketPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model := plan.upsert(ctx, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *bucketPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BucketPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model := plan.upsert(ctx, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *bucketPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BucketPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model := state.read(r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *bucketPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BucketPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := state.delete(r.client)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
}

func (r *bucketPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	model := BucketPolicyResourceModel{
		BucketName: types.StringValue(req.ID),
	}

	state := model.read(r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
