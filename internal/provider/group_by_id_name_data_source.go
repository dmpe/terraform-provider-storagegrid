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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &groupDataSource{}
var _ datasource.DataSourceWithConfigure = &groupDataSource{}

func NewGroupDataSource() datasource.DataSource {
	return &groupDataSource{}
}

// groupDataSource defines the data source implementation.
type groupDataSource struct {
	client *S3GridClient
}

func (d *groupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *groupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Fetch a specific group - a data source",
		Attributes: map[string]schema.Attribute{
			id: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(
						path.MatchRelative().AtParent().AtName(id),
						path.MatchRelative().AtParent().AtName(unique_name),
					),
				},
			},
			"account_id": schema.StringAttribute{
				Computed: true,
			},
			"display_name": schema.StringAttribute{
				Computed: true,
			},
			unique_name: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(
						path.MatchRelative().AtParent().AtName(id),
						path.MatchRelative().AtParent().AtName(unique_name),
					),
				},
			},
			"group_urn": schema.StringAttribute{
				Computed: true,
			},
			"federated": schema.BoolAttribute{
				Computed: true,
			},
			"management_read_only": schema.BoolAttribute{
				Computed: true,
			},
			"policies": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"management": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"manage_all_containers": schema.BoolAttribute{
								Computed:    true,
								Description: "Ability to manage all S3 buckets or Swift containers for this tenant account (overrides permission settings in group or bucket policies). Supersedes the viewAllContainers permission",
							},
							"manage_endpoints": schema.BoolAttribute{
								Computed:    true,
								Description: "Allows users to configure endpoints for platform services.",
							},
							"manage_own_container_objects": schema.BoolAttribute{
								Computed:    true,
								Description: "Ability to use S3 Console to view and manage bucket objects",
							},
							"manage_own_s3_credentials": schema.BoolAttribute{
								Computed:    true,
								Description: "Allows users to create and delete their own S3 access keys.",
							},
							"view_all_containers": schema.BoolAttribute{
								Computed:    true,
								Description: "Allows users to view settings of all S3 buckets (or Swift containers) in this account. Superseded by the Manage all buckets permission. Applies to the Tenant Manager UI and API only and does not affect the permissions granted by an S3 group policy.",
							},
							"root_access": schema.BoolAttribute{
								Computed:    true,
								Description: "Allows users to access all administration features. Root access permission supersedes all other permissions.",
							},
						},
					},
					"s3": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							id: schema.StringAttribute{
								Optional:            true,
								Computed:            true,
								Description:         "S3 Policy ID provided by policy generator tools (currently not used)",
								MarkdownDescription: "S3 Policy ID provided by policy generator tools (currently not used)",
							},
							"statement": schema.ListNestedAttribute{
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										act: schema.ListAttribute{
											ElementType:         types.StringType,
											Optional:            true,
											Computed:            true,
											Description:         "the specific actions that will be allowed (Can be a string if only one element. A statement must have either Action or NotAction.)",
											MarkdownDescription: "the specific actions that will be allowed (Can be a string if only one element. A statement must have either Action or NotAction.)",
										},
										"effect": schema.StringAttribute{
											Optional:            true,
											Computed:            true,
											Description:         "the specific result of the statement (either an allow or an explicit deny)",
											MarkdownDescription: "the specific result of the statement (either an allow or an explicit deny)",
											Validators: []validator.String{
												stringvalidator.OneOf(
													"Allow",
													"Deny",
												),
											},
										},
										n_act: schema.ListAttribute{
											ElementType:         types.StringType,
											Optional:            true,
											Computed:            true,
											Description:         "the specific exceptional actions (Can be a string if only one element. A statement must have either Action or NotAction.)",
											MarkdownDescription: "the specific exceptional actions (Can be a string if only one element. A statement must have either Action or NotAction.)",
										},
										n_res: schema.ListAttribute{
											ElementType:         types.StringType,
											Optional:            true,
											Computed:            true,
											Description:         "the objects that the statement does not cover (Can be a string if only one element. A statement must have either Resource or NotResource.)",
											MarkdownDescription: "the objects that the statement does not cover (Can be a string if only one element. A statement must have either Resource or NotResource.)",
										},
										res: schema.ListAttribute{
											ElementType:         types.StringType,
											Optional:            true,
											Computed:            true,
											Description:         "the objects that the statement covers (Can be a string if only one element. A statement must have either Resource or NotResource.)",
											MarkdownDescription: "the objects that the statement covers (Can be a string if only one element. A statement must have either Resource or NotResource.)",
										},
										"sid": schema.StringAttribute{
											Optional:            true,
											Computed:            true,
											Description:         "an optional identifier that you provide for the policy statement",
											MarkdownDescription: "an optional identifier that you provide for the policy statement",
										},
									},
								},
								Required:            true,
								Description:         "a list of group policy statements",
								MarkdownDescription: "a list of group policy statements",
							},
							"version": schema.StringAttribute{
								Optional:            true,
								Computed:            true,
								Description:         "S3 API Version (currently not used)",
								MarkdownDescription: "S3 API Version (currently not used)",
							},
						},
					},
				},
			},
		},
	}
}

func (d *groupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *groupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state GroupsDataSourceModel
	var jsonData groupsDataSourceGolangModelSingle
	var s3Sts []*S3PolicyStatementDataModel
	var idType types.String
	var uniqueNameType types.String
	var fullPath string

	// Read Terraform configuration data into the model
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("id"), &idType)...)

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("unique_name"), &uniqueNameType)...)

	tflog.Debug(ctx, "1. Fetching group by id or by name.")
	if !state.ID.IsNull() {
		fullPath = api_groups + "/" + idType.ValueString()
	} else {
		fullPath = api_groups + "/" + uniqueNameType.ValueString()
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
	mgmtPolicies := &ManagementPolicyDataModel{
		ManageAllContainers:       types.BoolValue(jsonData.Data.Policies.Management.ManageAllContainers),
		ManageEndpoints:           types.BoolValue(jsonData.Data.Policies.Management.ManageEndpoints),
		ManageOwnContainerObjects: types.BoolValue(jsonData.Data.Policies.Management.ManageOwnContainerObjects),
		ManageOwnS3Credentials:    types.BoolValue(jsonData.Data.Policies.Management.ManageOwnS3Credentials),
		ViewAllContainers:         types.BoolValue(jsonData.Data.Policies.Management.ViewAllContainers),
		RootAccess:                types.BoolValue(jsonData.Data.Policies.Management.RootAccess),
	}
	for _, s3Pol := range jsonData.Data.Policies.S3.Statement {
		actions := []types.String{}
		notActions := []types.String{}
		resources := []types.String{}
		notResources := []types.String{}

		for _, action := range s3Pol.Action.AsStringSlice() {
			actions = append(actions, types.StringValue(action))
		}
		for _, notAction := range s3Pol.NotAction.AsStringSlice() {
			notActions = append(notActions, types.StringValue(notAction))
		}
		for _, resource := range s3Pol.Resource.AsStringSlice() {
			resources = append(resources, types.StringValue(resource))
		}
		for _, notResource := range s3Pol.NotResource.AsStringSlice() {
			notResources = append(notResources, types.StringValue(notResource))
		}
		s3Statement := &S3PolicyStatementDataModel{
			Sid:         types.StringValue(s3Pol.Sid),
			Effect:      types.StringValue(s3Pol.Effect),
			Action:      actions,
			NotAction:   notActions,
			Resource:    resources,
			NotResource: notResources,
		}
		s3Sts = append(s3Sts, s3Statement)
	}

	s3Policy := &S3PolicyDataModel{
		ID:        types.StringValue(jsonData.Data.Policies.S3.ID),
		Version:   types.StringValue(jsonData.Data.Policies.S3.Version),
		Statement: s3Sts,
	}

	groupDataSingle := &GroupsDataSourceModel{
		ID:                 types.StringValue(jsonData.Data.ID),
		AccountID:          types.StringValue(jsonData.Data.AccountID),
		DisplayName:        types.StringValue(jsonData.Data.DisplayName),
		UniqueName:         types.StringValue(jsonData.Data.UniqueName),
		GroupURN:           types.StringValue(jsonData.Data.GroupURN),
		Federated:          types.BoolValue(jsonData.Data.Federated),
		ManagementReadOnly: types.BoolValue(jsonData.Data.ManagementReadOnly),
		Policies: &PoliciesModel{
			Management: mgmtPolicies,
			S3:         s3Policy,
		},
	}
	state = *groupDataSingle

	resp.Diagnostics.Append(diags...)
	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "finihed read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
